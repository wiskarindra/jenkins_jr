package jenkins_jr

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bukalapak/apinizer/response"
	"github.com/bukalapak/jenkins_jr/config"
	"github.com/bukalapak/jenkins_jr/pkg/api"
	"github.com/bukalapak/jenkins_jr/pkg/currentuser"
	"github.com/bukalapak/jenkins_jr/pkg/log"
)

// PostEntity represents posts entity.
type PostEntity struct {
	ID             int64  `db:"id"`
	Title          string `db:"title"`
	Description    string `db:"description"`
	InfluencerName string `db:"influencer_name"`
	Published      bool   `db:"published"`
	LikeCount      int    `db:"like_count"`
	Deleted        bool   `db:"deleted"`
	InfluencerID   int64  `db:"influencer_id"`

	FirstPublishedAt *time.Time `db:"first_published_at"`
	LastPublishedAt  *time.Time `db:"last_published_at"`
	CreatedAt        *time.Time `db:"created_at"`
	UpdatedAt        *time.Time `db:"updated_at"`

	Images []PostImageEntity
}

// PostRequest represents post in requests.
type PostRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	InfluencerName string `json:"influencer_name"`
	Published      bool   `json:"published"`

	Images []PostImageRequest `json:"images"`
}

// PostExclusiveResponse represents response to ExclusiveIndex.
type PostExclusiveResponse struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	InfluencerName string `json:"influencer_name"`
	Published      bool   `json:"published"`
	LikeCount      int    `json:"like_count"`

	LastPublishedAt string `json:"last_published_at"`
	ModifiedAt      string `json:"modified_at"`

	Images []PostImageResponse `json:"images"`
}

// PostPublicResponse represents response to PublicIndex.
type PostPublicResponse struct {
	ID             int64  `json:"id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	InfluencerName string `json:"influencer_name"`
	URL            string `json:"url"`

	PublishedAt string `json:"published_at"`

	Likes  PostLikeResponse    `json:"likes"`
	Images []PostImageResponse `json:"images"`

	Influencer InfluencerResponse `json:"influencer"`

	// TODO: deprecate soon
	Image *PostImageResponse `json:"image"`
	Tags  []PostTagResponse  `json:"tags"`
}

// PublishResponse represents response to Publish.
type PublishResponse struct {
	Published bool `json:"published"`
}

// MessageResponse represents response message.
type MessageResponse struct {
	Message string      `json:"message"`
	Meta    interface{} `json:"meta"`
}

const (
	postSelectQuery = "SELECT id, title, description, influencer_name, published, like_count, first_published_at, last_published_at, created_at, updated_at, deleted, influencer_id FROM posts "
)

// PublicIndex sends as response posts for non-admin users.
func (env *Env) PublicIndex(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	params := r.URL.Query()
	search := env.buildPublicPostSearchForm(params)

	if filter, ok := env.buildFilterForm(ctx, params); ok {
		filterEntities, err := env.filterPost(ctx, filter)
		if err != nil {
			response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
			return err
		}
		search.applyFilters(filterEntities)
	}

	postEntities, err := env.searchPost(ctx, search)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	meta := api.IndexMeta{HTTPStatus: http.StatusOK, Limit: search.limit, Offset: search.offset, Total: len(postEntities)}
	if len(postEntities) == 0 {
		response.Write(w, response.BuildSuccess([]PostPublicResponse{}, "", meta), http.StatusOK)
		return nil
	}

	postIDs := getPostIDsFromPosts(postEntities)
	images, err := env.findImagesByPostIDs(ctx, postIDs)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}
	tags, err := env.findTagsByPostIDs(ctx, postIDs)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}
	mergeTagsToImages(images, tags)

	likedPostIDs, err := env.findLikedPostIDs(ctx, postIDs)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	influencerIDs := getInfluencerIDsFromPosts(postEntities)
	influencers, err := env.selectInfluencersByIDs(ctx, influencerIDs)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	postResponses := buildPublicPostResponses(ctx, postEntities, images, tags, likedPostIDs, influencers)
	response.Write(w, response.BuildSuccess(postResponses, "", meta), http.StatusOK)
	return nil
}

// ExclusiveIndex sends as response posts for admins.
func (env *Env) ExclusiveIndex(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	params := r.URL.Query()
	search := env.buildExclusivePostSearchForm(ctx, params)

	if filter, ok := env.buildFilterForm(ctx, params); ok {
		filterEntities, err := env.filterPost(ctx, filter)
		if err != nil {
			response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
			return err
		}
		search.applyFilters(filterEntities)
	}

	postEntities, err := env.searchPost(ctx, search)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	meta := api.IndexMeta{HTTPStatus: http.StatusOK, Limit: search.limit, Offset: search.offset, Total: len(postEntities), TotalPages: env.countPostTotalPage(ctx, search)}
	if len(postEntities) == 0 {
		response.Write(w, response.BuildSuccess([]PostExclusiveResponse{}, "", meta), http.StatusOK)
		return nil
	}

	postIDs := getPostIDsFromPosts(postEntities)
	images, err := env.findImagesByPostIDs(ctx, postIDs)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}
	tags, err := env.findTagsByPostIDs(ctx, postIDs)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}
	mergeTagsToImages(images, tags)

	postResponses := buildExclusivePostResponses(ctx, postEntities, images, tags)
	response.Write(w, response.BuildSuccess(postResponses, "", meta), http.StatusOK)
	return nil
}

// Show sends as response details of requested post.
func (env *Env) Show(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	postID, err := parseIntContext(ctx, api.GetParams(r).ByName("id"))
	if err != nil {
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}

	post, err := env.findPost(ctx, postID)
	if err != nil || post.Deleted {
		response.Write(w, response.BuildError(api.PostNotFoundError), api.PostNotFoundError.HTTPCode)
		return err
	}

	images, _ := env.findImagesByPostID(ctx, postID)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	tags, _ := env.findTagsByPostID(ctx, postID)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}
	for i := range tags {
		// to give BL admins simpler life
		tags[i].URL = strings.Split(tags[i].URL, "?")[0]
	}
	mergeTagsToImages(images, tags)

	publishedAtJSON := ""
	if post.LastPublishedAt != nil {
		publishedAtJSON = post.LastPublishedAt.Format(config.ResponseDatetimeFormat)
	}

	p := PostExclusiveResponse{
		ID:             post.ID,
		Title:          post.Title,
		Description:    post.Description,
		InfluencerName: post.InfluencerName,
		Published:      post.Published,
		LikeCount:      post.LikeCount,

		LastPublishedAt: publishedAtJSON,
		ModifiedAt:      post.UpdatedAt.Format(config.ResponseDatetimeFormat),

		Images: buildImagesResponse(images),
	}

	m := response.MetaInfo{HTTPStatus: http.StatusOK}
	response.Write(w, response.BuildSuccess(p, "", m), http.StatusOK)
	return nil
}

// Create creates new post.
func (env *Env) Create(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	var pr PostRequest
	err := json.NewDecoder(r.Body).Decode(&pr)
	if err != nil || !pr.isValid() {
		log.ErrLog(ctx, err, "decode-json", "create post fail")
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}
	defer r.Body.Close()

	createdAtDB := timeNowDB()
	createdAtJSON := timeNowResponse()

	tx := env.DB.MustBegin()

	influencer, err := env.findOrCreateInfluencer(ctx, tx, pr.InfluencerName)
	if err != nil {
		tx.Rollback()
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	postInsertValues := map[string]interface{}{
		"title":           pr.Title,
		"description":     pr.Description,
		"influencer_id":   influencer.ID,
		"influencer_name": influencer.Name,
		"published":       pr.Published,
		"created_at":      createdAtDB,
		"updated_at":      createdAtDB,
	}

	publishedAtJSON := ""
	if pr.Published {
		postInsertValues["first_published_at"] = createdAtDB
		postInsertValues["last_published_at"] = createdAtDB
		publishedAtJSON = createdAtJSON
	}

	keys := reflectValuesToString(reflect.ValueOf(postInsertValues).MapKeys())
	postInsertQuery := fmt.Sprintf("INSERT INTO posts (%s) VALUES (:%s)", strings.Join(keys, ", "), strings.Join(keys, ", :"))

	row, err := tx.NamedExecContext(ctx, postInsertQuery, postInsertValues)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "insert-sql-query", "insert to posts fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}
	postID, _ := row.LastInsertId()

	for pos, image := range pr.Images {
		var imageID int64
		if image.URL != "" {
			row, err := addImage(ctx, tx, postID, image, pos)
			if err != nil {
				response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
				return err
			}
			imageID, _ = row.LastInsertId()
		}

		if len(image.Tags) > 0 {
			tagInsertQuery := "INSERT INTO post_tags (post_id, post_image_id, name, url, coord_x, coord_y, created_at, updated_at, reference_id, reference_type, bukalapak_category_id) VALUES "
			tagInsertValues := []interface{}{}

			for _, tag := range image.Tags {
				var bukalapakCategoryID int64
				if tag.Category != (PostTagCategoryRequest{}) {
					bukalapakCategoryID = tag.Category.ID
					addCategory(ctx, tx, tag)

					if pr.Published {
						err := incrementCategoryCount(ctx, tx, bukalapakCategoryID)
						if err != nil {
							return err
						}
					}
				}

				tagURL, err := url.Parse(tag.URL)
				if err != nil {
					tx.Rollback()
					log.ErrLog(ctx, err, "parse-tag-url", "parse post_tags url fail")
				}
				refID, err := strconv.ParseInt(tagURL.Query().Get("ref_id"), 36, 64)
				if err != nil {
					log.ErrLog(ctx, err, "parse-tag-url", "parse post_tags ref_id fail")
				}
				refType := tagURL.Query().Get("ref_type")

				tagInsertQuery += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
				tagInsertValues = append(tagInsertValues, postID, imageID, tag.Name, tag.URL, tag.Coord.X, tag.Coord.Y, createdAtDB, createdAtDB, refID, refType, bukalapakCategoryID)

				if bukalapakCategoryID != 0 {
					filterInsertQuery := "INSERT INTO post_filters (post_id, bukalapak_category_id) VALUES (:post_id, :bukalapak_category_id)"
					filterInsertValues := map[string]interface{}{
						"post_id":               postID,
						"bukalapak_category_id": bukalapakCategoryID,
					}
					tx.NamedExecContext(ctx, filterInsertQuery, filterInsertValues)
				}
			}

			// trim the last ','
			tagInsertQuery = tagInsertQuery[0 : len(tagInsertQuery)-1]

			stmt, err := tx.Prepare(tagInsertQuery)
			if err != nil {
				tx.Rollback()
				log.ErrLog(ctx, err, "prepare-sql-query", "prepare tag insert query fail")
				response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
				return err
			}
			if _, err := stmt.ExecContext(ctx, tagInsertValues...); err != nil {
				tx.Rollback()
				log.ErrLog(ctx, err, "insert-sql-query", "insert to post_tags fail")
				response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.ErrLog(ctx, err, "commit-sql-tx", "create post commit fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	// TODO: use goroutine
	env.addPostLogHistory(ctx, PostHistoryData{PostID: postID, Was: PostEntity{}})

	p := PostExclusiveResponse{
		ID:             postID,
		Title:          pr.Title,
		Description:    pr.Description,
		InfluencerName: influencer.Name,
		Published:      pr.Published,
		LikeCount:      0,

		LastPublishedAt: publishedAtJSON,
		ModifiedAt:      createdAtJSON,

		Images: convertImagesRequestToResponse(pr.Images),
	}

	m := response.MetaInfo{HTTPStatus: http.StatusCreated}
	response.Write(w, response.BuildSuccess(p, "", m), http.StatusCreated)
	return nil
}

// Update updates requested post.
func (env *Env) Update(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	postID, err := parseIntContext(ctx, api.GetParams(r).ByName("id"))
	if err != nil {
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}

	var pr PostRequest
	err = json.NewDecoder(r.Body).Decode(&pr)
	if err != nil || !pr.isValid() {
		log.ErrLog(ctx, err, "decode-json", "update post fail")
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}
	defer r.Body.Close()

	post, err := env.findPost(ctx, postID)
	if err != nil || post.Deleted {
		response.Write(w, response.BuildError(api.PostNotFoundError), api.PostNotFoundError.HTTPCode)
		return err
	}

	updatedAtDB := timeNowDB()
	updatedAtJSON := timeNowResponse()

	tx := env.DB.MustBegin()

	influencer, err := env.findOrCreateInfluencer(ctx, tx, pr.InfluencerName)
	if err != nil {
		tx.Rollback()
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	if post.Published {
		err := decrementCategoryCounts(ctx, tx, post)
		if err != nil {
			return err
		}
	}

	var updateSetQuery []string
	for _, k := range []string{"title", "description", "influencer_id", "influencer_name", "published", "updated_at"} {
		updateSetQuery = append(updateSetQuery, fmt.Sprintf("%s=:%s", k, k))
	}

	var publishedAtJSON string
	if post.LastPublishedAt == nil { // NEVER PUBLISHED
		if pr.Published {
			updateSetQuery = append(updateSetQuery, []string{"first_published_at=:updated_at", "last_published_at=:updated_at"}...)
			publishedAtJSON = updatedAtJSON
		}
	} else if post.LastPublishedAt != nil && post.Published { // HAVE PUBLISHED
		publishedAtJSON = post.LastPublishedAt.Format(config.ResponseDatetimeFormat)
	} else if post.LastPublishedAt != nil && !post.Published { // WAS PUBLISHED
		if pr.Published {
			updateSetQuery = append(updateSetQuery, "last_published_at=:updated_at")
			publishedAtJSON = updatedAtJSON
		} else {
			publishedAtJSON = post.LastPublishedAt.Format(config.ResponseDatetimeFormat)
		}
	}

	updatePostQuery := fmt.Sprintf("UPDATE posts SET %s WHERE id=:id", strings.Join(updateSetQuery, ", "))
	updateSetValue := map[string]interface{}{
		"id":              postID,
		"title":           pr.Title,
		"description":     pr.Description,
		"influencer_id":   influencer.ID,
		"influencer_name": influencer.Name,
		"published":       pr.Published,
		"updated_at":      updatedAtDB,
	}
	_, err = tx.NamedExecContext(ctx, updatePostQuery, updateSetValue)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "update-sql-query", "update posts fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM post_images WHERE post_id=?", postID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "delete-sql-query", "delete post_images fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM post_tags WHERE post_id=?", postID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "delete-sql-query", "delete post_tags fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM post_filters WHERE post_id=?", postID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "delete-sql-query", "delete post_tags fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	for pos, image := range pr.Images {
		var imageID int64
		if image.URL != "" {
			row, err := addImage(ctx, tx, postID, image, pos)
			if err != nil {
				response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
				return err
			}
			imageID, _ = row.LastInsertId()
		}

		if len(image.Tags) > 0 {
			tagInsertQuery := "INSERT INTO post_tags (post_id, post_image_id, name, url, coord_x, coord_y, created_at, updated_at, reference_id, reference_type, bukalapak_category_id) VALUES "
			tagInsertValues := []interface{}{}

			for _, tag := range image.Tags {
				var bukalapakCategoryID int64
				if tag.Category != (PostTagCategoryRequest{}) {
					bukalapakCategoryID = tag.Category.ID
					addCategory(ctx, tx, tag)

					if pr.Published {
						err := incrementCategoryCount(ctx, tx, bukalapakCategoryID)
						if err != nil {
							return err
						}
					}
				}

				tagURL, err := url.Parse(tag.URL)
				if err != nil {
					tx.Rollback()
					log.ErrLog(ctx, err, "parse-tag-url", "parse post_tags url fail")
				}
				refID, err := strconv.ParseInt(tagURL.Query().Get("ref_id"), 36, 64)
				if err != nil {
					log.ErrLog(ctx, err, "parse-tag-url", "parse post_tags ref_id fail")
				}
				refType := tagURL.Query().Get("ref_type")

				tagInsertQuery += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
				tagInsertValues = append(tagInsertValues, postID, imageID, tag.Name, tag.URL, tag.Coord.X, tag.Coord.Y, updatedAtDB, updatedAtDB, refID, refType, bukalapakCategoryID)

				if bukalapakCategoryID != 0 {
					filterInsertQuery := "INSERT INTO post_filters (post_id, bukalapak_category_id) VALUES (:post_id, :bukalapak_category_id)"
					filterInsertValues := map[string]interface{}{
						"post_id":               postID,
						"bukalapak_category_id": tag.Category.ID,
					}
					tx.NamedExecContext(ctx, filterInsertQuery, filterInsertValues)
				}
			}

			// trim the last ','
			tagInsertQuery = tagInsertQuery[0 : len(tagInsertQuery)-1]

			stmt, err := tx.Prepare(tagInsertQuery)
			if err != nil {
				tx.Rollback()
				log.ErrLog(ctx, err, "prepare-sql-query", "prepare tag insert query fail")
				response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
				return err
			}
			if _, err := stmt.ExecContext(ctx, tagInsertValues...); err != nil {
				tx.Rollback()
				log.ErrLog(ctx, err, "insert-sql-query", "insert to post_tags fail")
				response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.ErrLog(ctx, err, "commit-sql-tx", "update post commit fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	log.InfoLog(ctx, fmt.Sprintf("%+v", pr), "update")

	// TODO: use goroutine
	env.addPostLogHistory(ctx, PostHistoryData{PostID: postID, Was: post})

	p := PostExclusiveResponse{
		ID:             postID,
		Title:          pr.Title,
		Description:    pr.Description,
		InfluencerName: influencer.Name,
		Published:      pr.Published,
		LikeCount:      post.LikeCount,

		LastPublishedAt: publishedAtJSON,
		ModifiedAt:      updatedAtJSON,

		Images: convertImagesRequestToResponse(pr.Images),
	}
	m := response.MetaInfo{HTTPStatus: http.StatusAccepted}
	response.Write(w, response.BuildSuccess(p, "", m), http.StatusAccepted)
	return nil
}

// Publish publishes/unpublishes requested post.
func (env *Env) Publish(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	postID, err := parseIntContext(ctx, api.GetParams(r).ByName("id"))
	if err != nil {
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}

	post, err := env.findPost(ctx, postID)
	if err != nil || post.Deleted {
		response.Write(w, response.BuildError(api.PostNotFoundError), api.PostNotFoundError.HTTPCode)
		return err
	}

	var pr struct {
		Published bool `json:"published"`
	}

	err = json.NewDecoder(r.Body).Decode(&pr)
	if err != nil {
		log.ErrLog(ctx, err, "decode-json", "publish post fail")
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}
	defer r.Body.Close()

	updatedAtDB := timeNowDB()
	publishValues := map[string]interface{}{
		"post_id":    postID,
		"published":  pr.Published,
		"updated_at": updatedAtDB,
	}

	tx := env.DB.MustBegin()

	publishQuery := "UPDATE posts SET published=:published WHERE id=:post_id"
	if post.LastPublishedAt == nil { // NEVER PUBLISHED
		if pr.Published {
			publishQuery = "UPDATE posts SET published=:published, first_published_at=:updated_at, last_published_at=:updated_at, updated_at=:updated_at WHERE id=:post_id"

			err := incrementCategoryCounts(ctx, tx, post)
			if err != nil {
				return err
			}
		}
	} else if post.LastPublishedAt != nil && post.Published { // HAVE PUBLISHED
		if !pr.Published {
			publishQuery = "UPDATE posts SET published=:published, updated_at=:updated_at WHERE id=:post_id"

			err := decrementCategoryCounts(ctx, tx, post)
			if err != nil {
				return err
			}
		}
	} else if post.LastPublishedAt != nil && !post.Published { // WAS PUBLISHED
		if pr.Published {
			publishQuery = "UPDATE posts SET published=:published, last_published_at=:updated_at, updated_at=:updated_at WHERE id=:post_id"

			err := incrementCategoryCounts(ctx, tx, post)
			if err != nil {
				return err
			}
		}
	}

	_, err = tx.NamedExecContext(ctx, publishQuery, publishValues)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "update-sql-query", "update posts published fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.ErrLog(ctx, err, "commit-sql-tx", "create post commit fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	// TODO: use goroutine
	env.addPostLogHistory(ctx, PostHistoryData{PostID: postID, Was: post})

	m := response.MetaInfo{HTTPStatus: http.StatusAccepted}
	response.Write(w, response.BuildSuccess(PublishResponse{pr.Published}, "", m), http.StatusAccepted)
	return nil
}

// Delete marks requested post as deleted.
func (env *Env) Delete(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	postID, err := parseIntContext(ctx, api.GetParams(r).ByName("id"))
	if err != nil {
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}

	post, err := env.findPost(ctx, postID)
	if err != nil || post.Deleted {
		response.Write(w, response.BuildError(api.PostNotFoundError), api.PostNotFoundError.HTTPCode)
		return err
	}

	tx := env.DB.MustBegin()

	updatedAt := time.Now()
	_, err = tx.ExecContext(ctx, "UPDATE posts SET deleted=true, published=false, updated_at=? WHERE id=?", timeToStringDB(&updatedAt), postID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "update-sql-query", "update posts deleted fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	if post.Published {
		err := decrementCategoryCounts(ctx, tx, post)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM post_filters WHERE post_id=?", postID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "delete-sql-query", "delete post_tags fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.ErrLog(ctx, err, "commit-sql-tx", "create post commit fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	// TODO: use goroutine
	env.addPostLogHistory(ctx, PostHistoryData{PostID: postID, Was: post})

	m := response.MetaInfo{HTTPStatus: http.StatusAccepted}
	response.Write(w, MessageResponse{"Post has been deleted", m}, http.StatusAccepted)
	return nil
}

func (env *Env) findPost(ctx context.Context, postID int64) (PostEntity, error) {
	var post PostEntity
	switch err := env.DB.GetContext(ctx, &post, postSelectQuery+"WHERE id=? LIMIT 1", postID); err {
	case sql.ErrNoRows:
		return PostEntity{}, api.PostNotFoundError

	case nil:
		images, err := env.findImagesByPostID(ctx, postID)
		if err != nil {
			return PostEntity{}, err
		}
		tags, err := env.findTagsByPostID(ctx, postID)
		if err != nil {
			return PostEntity{}, err
		}
		mergeTagsToImages(images, tags)

		post.Images = images
		return post, nil

	default:
		log.ErrLog(ctx, err, "select-sql-query", "select from posts fail")
		return post, err
	}
}

func (env *Env) isPostPresent(ctx context.Context, postID int64) bool {
	var present bool
	env.DB.GetContext(ctx, &present, "SELECT 1 AS one FROM posts WHERE id=? LIMIT 1", postID)
	return present
}

func (pr PostRequest) isValid() bool {
	for _, image := range pr.Images {
		if image.URL == "" || image.Height == 0 || image.Width == 0 || len(image.Tags) > 10 {
			return false
		}
	}
	return true
}

func buildPublicPostResponses(ctx context.Context, posts []PostEntity, images []PostImageEntity, tags []PostTagEntity, likedPostIDs []int64, influencers []InfluencerEntity) []PostPublicResponse {
	im := mapPostIDToImages(images)
	ifm := mapInfluencerIDToInfluencers(influencers)
	currentUser := currentuser.FromContext(ctx)

	var postResponses []PostPublicResponse
	for _, post := range posts {
		images := im[post.ID]

		liked := isInSliceInt64(post.ID, likedPostIDs)

		// TODO: deprecate soon
		var firstImage *PostImageResponse
		var firstImageTags []PostTagResponse
		if len(images) > 0 {
			firstImageTags = buildTagsResponse(images[0].Tags)
			firstImage = &PostImageResponse{
				URL:    images[0].styleURL(config.IndexImageURLStyle),
				Height: images[0].Height,
				Width:  images[0].Width,
				Tags:   firstImageTags,
			}
		}

		imagesResponse := buildImagesResponseWithStyle(images, config.IndexImageURLStyle)
		for i := range imagesResponse {
			for j := range imagesResponse[i].Tags {
				if tag := imagesResponse[i].Tags[j]; strings.Contains(tag.URL, "?") {
					imagesResponse[i].Tags[j].URL = fmt.Sprintf("%s&from=inspiration&infn=%s", tag.URL, post.InfluencerName)
				} else {
					imagesResponse[i].Tags[j].URL = fmt.Sprintf("%s?from=inspiration&infn=%s", tag.URL, post.InfluencerName)
				}
			}
		}

		postResponses = append(postResponses, PostPublicResponse{
			ID:             post.ID,
			Title:          post.Title,
			Description:    post.getDescription(currentUser),
			InfluencerName: post.InfluencerName,
			URL:            fmt.Sprintf("%s#%d", config.InspirationIndexURL(), post.ID),

			PublishedAt: post.FirstPublishedAt.Format(config.ResponseDatetimeFormat),

			Influencer: buildInfluencerResponse(ifm[post.InfluencerID]),
			Likes:      PostLikeResponse{UserLike: liked, Count: normalizeLikeCount(ctx, post.LikeCount)},
			Images:     imagesResponse,

			// TODO: deprecate soon
			Image: firstImage,
			Tags:  firstImageTags,
		})
	}
	return postResponses
}

func buildExclusivePostResponses(ctx context.Context, posts []PostEntity, images []PostImageEntity, tags []PostTagEntity) []PostExclusiveResponse {
	im := mapPostIDToImages(images)

	postResponses := []PostExclusiveResponse{}
	for _, post := range posts {
		images := im[post.ID]

		publishedAtJSON := ""
		if post.LastPublishedAt != nil {
			publishedAtJSON = post.LastPublishedAt.Format(config.ResponseDatetimeFormat)
		}

		postResponses = append(postResponses, PostExclusiveResponse{
			ID:             post.ID,
			Title:          post.Title,
			Description:    post.Description,
			InfluencerName: post.InfluencerName,
			Published:      post.Published,
			LikeCount:      normalizeLikeCount(ctx, post.LikeCount),

			LastPublishedAt: publishedAtJSON,
			ModifiedAt:      post.UpdatedAt.Format(config.ResponseDatetimeFormat),

			Images: buildImagesResponse(images),
		})
	}
	return postResponses
}

func getPostIDsFromPosts(posts []PostEntity) []int64 {
	postIDs := make([]int64, 0)
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
	}

	return postIDs
}

func (pe PostEntity) getDescription(currentUser *currentuser.CurrentUser) string {
	// Remove all html tag for android and ios old version
	if (currentUser.Platform() == "blios" && currentUser.AppVersion() < 224000) || (currentUser.Platform() == "blandroid" && currentUser.AppVersion() < 4029000) {
		regexNewLine := regexp.MustCompile(`<br/>|</[^>]*><div>`)
		description := regexNewLine.ReplaceAllString(pe.Description, "\n")
		regexTag := regexp.MustCompile(`<\/?[^>]*>`)
		return regexTag.ReplaceAllString(description, "")
	}

	// Change html tag newline to <br/> for android version support description link
	if currentUser.Platform() == "blandroid" {
		re := regexp.MustCompile(`</[^>]*><span>`)
		description := re.ReplaceAllString(pe.Description, "")
		re = regexp.MustCompile(`</[^>]*><div>`)
		return re.ReplaceAllString(description, "<br/>")
	}
	return pe.Description
}

func getInfluencerIDsFromPosts(posts []PostEntity) []int64 {
	influencerIDs := make([]int64, 0)
	for _, post := range posts {
		influencerIDs = append(influencerIDs, post.InfluencerID)
	}
	return influencerIDs
}
