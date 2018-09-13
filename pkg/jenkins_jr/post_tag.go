package jenkins_jr

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/bukalapak/jenkins_jr/pkg/log"
)

// PostTagEntity represents post_tags entity.
type PostTagEntity struct {
	PostID              int64   `db:"post_id"`
	ImageID             int64   `db:"post_image_id"`
	Name                string  `db:"name"`
	URL                 string  `db:"url"`
	CoordX              float32 `db:"coord_x"`
	CoordY              float32 `db:"coord_y"`
	BukalapakCategoryID int64   `db:"bukalapak_category_id"`
}

// PostTagRequest represents tag in requests.
type PostTagRequest struct {
	Name     string                 `json:"name"`
	URL      string                 `json:"url"`
	Category PostTagCategoryRequest `json:"category"`
	Coord    PostTagCoordRequest    `json:"coord"`
}

// PostTagResponse represents tag in responses.
type PostTagResponse struct {
	Name  string               `json:"name"`
	URL   string               `json:"url"`
	Coord PostTagCoordResponse `json:"coord"`
}

// PostTagCoordRequest represents tag coordinate in requests.
type PostTagCoordRequest struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// PostTagCoordResponse represents tag coordinate in responses.
type PostTagCoordResponse struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

// PostTagCategoryRequest represents tag category in requests.
type PostTagCategoryRequest struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

const (
	postTagSelectQuery = "SELECT post_id, post_image_id, name, url, coord_x, coord_y, bukalapak_category_id FROM post_tags "
)

func (env *Env) findTagsByPostID(ctx context.Context, postID int64) ([]PostTagEntity, error) {
	condition := fmt.Sprintf("WHERE post_id=%d LIMIT 10", postID)
	return env.findTags(ctx, condition)
}

func (env *Env) findTagsByPostIDs(ctx context.Context, postIDs []int64) ([]PostTagEntity, error) {
	condition := fmt.Sprintf("WHERE post_id IN (%s)", joinSliceInt64(postIDs))
	return env.findTags(ctx, condition)
}

func (env *Env) findTags(ctx context.Context, condition string) ([]PostTagEntity, error) {
	query := postTagSelectQuery + condition

	var tags []PostTagEntity
	err := env.DB.SelectContext(ctx, &tags, query)
	if err != nil && err != sql.ErrNoRows {
		log.ErrLog(ctx, err, "select-sql-query", "select from post_tags fail")
		return nil, err
	}
	return tags, nil
}

func buildTagsResponse(tags []PostTagEntity) []PostTagResponse {
	trs := make([]PostTagResponse, 0)
	for _, tag := range tags {
		trs = append(trs, convertTagEntityToResponse(tag))
	}
	return trs
}

func convertTagEntityToResponse(tag PostTagEntity) PostTagResponse {
	return PostTagResponse{
		Name:  tag.Name,
		URL:   tag.URL,
		Coord: PostTagCoordResponse{tag.CoordX, tag.CoordY},
	}
}

func convertTagRequestToResponse(tag PostTagRequest) PostTagResponse {
	return PostTagResponse{
		Name: tag.Name,
		URL:  tag.URL,
		Coord: PostTagCoordResponse{
			tag.Coord.X,
			tag.Coord.Y,
		},
	}
}

func convertTagsRequestToResponse(tags []PostTagRequest) []PostTagResponse {
	resp := make([]PostTagResponse, 0)
	for _, tag := range tags {
		resp = append(resp, PostTagResponse{
			Name: tag.Name,
			URL:  tag.URL,
			Coord: PostTagCoordResponse{
				tag.Coord.X,
				tag.Coord.Y,
			},
		})
	}
	return resp
}

func mergeTagsToImages(images []PostImageEntity, tags []PostTagEntity) {
	tm := mapImageIDToTags(tags)
	for i, image := range images {
		images[i].Tags = tm[image.ID]
	}
}

func mapImageIDToTags(tags []PostTagEntity) map[int64][]PostTagEntity {
	m := make(map[int64][]PostTagEntity)
	for _, tag := range tags {
		m[tag.ImageID] = append(m[tag.ImageID], tag)
	}
	return m
}

func mapPostIDToTags(tags []PostTagEntity) map[int64][]PostTagEntity {
	m := make(map[int64][]PostTagEntity)
	for _, tag := range tags {
		m[tag.PostID] = append(m[tag.PostID], tag)
	}
	return m
}
