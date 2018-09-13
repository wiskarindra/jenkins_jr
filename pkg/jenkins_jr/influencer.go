package jenkins_jr

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/bukalapak/apinizer/response"
	"github.com/bukalapak/jenkins_jr/config"
	"github.com/bukalapak/jenkins_jr/pkg/api"
	"github.com/jmoiron/sqlx"
)

// InfluencerEntity represents influencers entity
type InfluencerEntity struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (ie InfluencerEntity) URL() string {
	return fmt.Sprintf("%s/%s", config.InfluencerIndexURL(), getInfluencerUsernameFromName(ie.Name))
}

type InfluencerResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

const (
	influencersSelectQuery = "SELECT id, name FROM influencers "
	influencersInsertQuery = "INSERT INTO influencers (name, created_at, updated_at) VALUES (:name, :created_at, :updated_at)"
)

func (env *Env) InfluencerShow(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	infID, err := parseIntContext(ctx, api.GetParams(r).ByName("id"))
	if err != nil {
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}

	influencer, err := env.findInfluencerByID(ctx, infID)
	if err != nil {
		response.Write(w, response.BuildError(api.InfluencerNotFoundError), api.InfluencerNotFoundError.HTTPCode)
		return err
	}

	m := response.MetaInfo{HTTPStatus: http.StatusOK}
	response.Write(w, response.BuildSuccess(buildInfluencerResponse(influencer), "", m), http.StatusOK)
	return nil
}

func (env *Env) InfluencerShowByUsername(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	name := getInfluencerNameFromUsername(api.GetParams(r).ByName("username"))
	influencer, err := env.findInfluencerByName(ctx, name)
	if err != nil {
		response.Write(w, response.BuildError(api.InfluencerNotFoundError), api.InfluencerNotFoundError.HTTPCode)
		return err
	}

	m := response.MetaInfo{HTTPStatus: http.StatusOK}
	response.Write(w, response.BuildSuccess(buildInfluencerResponse(influencer), "", m), http.StatusOK)
	return nil
}

func (env *Env) InfluencerPostIndex(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	infID, err := parseIntContext(ctx, api.GetParams(r).ByName("id"))
	if err != nil {
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}

	influencer, err := env.findInfluencerByID(ctx, infID)
	if err != nil {
		response.Write(w, response.BuildError(api.InfluencerNotFoundError), api.InfluencerNotFoundError.HTTPCode)
		return err
	}

	form := env.buildInfluencerIndexPostSearchForm(r.URL.Query(), influencer.ID)
	postEntities, err := env.searchPost(ctx, form)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	meta := api.IndexMeta{HTTPStatus: http.StatusOK, Limit: form.limit, Offset: form.offset, Total: len(postEntities)}
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

	postResponses := buildPublicPostResponses(ctx, postEntities, images, tags, likedPostIDs, []InfluencerEntity{influencer})
	response.Write(w, response.BuildSuccess(postResponses, "", meta), http.StatusOK)
	return nil
}

func (env *Env) InfluencerPostIndexByUsername(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	name := getInfluencerNameFromUsername(api.GetParams(r).ByName("username"))
	influencer, err := env.findInfluencerByName(ctx, name)
	if err != nil {
		response.Write(w, response.BuildError(api.InfluencerNotFoundError), api.InfluencerNotFoundError.HTTPCode)
		return err
	}

	form := env.buildInfluencerIndexPostSearchForm(r.URL.Query(), influencer.ID)
	postEntities, err := env.searchPost(ctx, form)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	meta := api.IndexMeta{HTTPStatus: http.StatusOK, Limit: form.limit, Offset: form.offset, Total: len(postEntities)}
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

	postResponses := buildPublicPostResponses(ctx, postEntities, images, tags, likedPostIDs, []InfluencerEntity{influencer})
	response.Write(w, response.BuildSuccess(postResponses, "", meta), http.StatusOK)
	return nil
}

func (env *Env) selectInfluencersByIDs(ctx context.Context, influencerIDs []int64) ([]InfluencerEntity, error) {
	influencers := []InfluencerEntity{}
	if err := env.DB.SelectContext(ctx, &influencers, influencersSelectQuery+fmt.Sprintf("WHERE id IN (%s)", joinSliceInt64(influencerIDs))); err != nil {
		return []InfluencerEntity{}, err
	}
	return influencers, nil
}

func (env *Env) findInfluencerByID(ctx context.Context, id int64) (InfluencerEntity, error) {
	influencer := InfluencerEntity{}
	if err := env.DB.GetContext(ctx, &influencer, influencersSelectQuery+" WHERE id=?", id); err != nil {
		return InfluencerEntity{}, err
	}
	return influencer, nil
}

func (env *Env) findInfluencerByName(ctx context.Context, name string) (InfluencerEntity, error) {
	name = normalizeInfluencerName(name)
	influencer := InfluencerEntity{}
	if err := env.DB.GetContext(ctx, &influencer, influencersSelectQuery+" WHERE name=?", name); err != nil {
		return InfluencerEntity{}, err
	}
	return influencer, nil
}

func (env *Env) createInfluencer(ctx context.Context, tx *sqlx.Tx, name string) (InfluencerEntity, error) {
	name = normalizeInfluencerName(name)
	row, err := tx.NamedExecContext(ctx, influencersInsertQuery, map[string]interface{}{
		"name":       name,
		"created_at": timeNowDB(),
		"updated_at": timeNowDB(),
	})

	if err != nil {
		return InfluencerEntity{}, err
	}

	influencer := InfluencerEntity{}
	influencer.ID, _ = row.LastInsertId()
	influencer.Name = name
	return influencer, nil
}

func (env *Env) findOrCreateInfluencer(ctx context.Context, tx *sqlx.Tx, name string) (InfluencerEntity, error) {
	influencer, err := env.findInfluencerByName(ctx, name)
	if err == sql.ErrNoRows {
		influencer, err = env.createInfluencer(ctx, tx, name)
	}

	if err != nil {
		// LOG HERE
		return InfluencerEntity{}, err
	}

	return influencer, nil
}

func getInfluencerNameFromUsername(username string) string {
	// replace dash with whitespace
	username = strings.Replace(username, "-", " ", -1)

	// random name => Random Name
	return normalizeInfluencerName(username)
}

func getInfluencerUsernameFromName(name string) string {
	// downcase the string
	name = strings.ToLower(name)

	// replace whitespace with "-"
	return strings.Replace(name, " ", "-", -1)
}

func normalizeInfluencerName(name string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
	name = reg.ReplaceAllString(name, "")

	// downcase the string
	name = strings.ToLower(name)

	// clear all whitespace
	name = strings.Join(strings.Fields(name), " ")

	// random name => Random Name
	return strings.Title(name)
}

func mapInfluencerIDToInfluencers(influencers []InfluencerEntity) map[int64]InfluencerEntity {
	m := make(map[int64]InfluencerEntity)
	for _, inf := range influencers {
		m[inf.ID] = inf
	}
	return m
}

func buildInfluencerResponse(in InfluencerEntity) InfluencerResponse {
	return InfluencerResponse{
		ID:   in.ID,
		Name: in.Name,
		URL:  in.URL(),
	}
}
