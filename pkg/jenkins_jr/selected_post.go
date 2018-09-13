package jenkins_jr

import (
	"net/http"

	"github.com/bukalapak/apinizer/response"
	"github.com/wiskarindra/jenkins_jr/pkg/api"
)

// SelectedPosts sends as response posts to put in selected posts section.
func (env *Env) SelectedPosts(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	form := env.buildSelectedPostsSearchForm(r.URL.Query())
	posts, err := env.searchPost(ctx, form)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	meta := api.IndexMeta{HTTPStatus: http.StatusOK, Limit: form.limit, Offset: form.offset, Total: len(posts)}
	if len(posts) == 0 {
		response.Write(w, response.BuildSuccess([]PostPublicResponse{}, "", meta), http.StatusOK)
		return nil
	}

	postIDs := getPostIDsFromPosts(posts)
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

	influencerIDs := getInfluencerIDsFromPosts(posts)
	influencers, err := env.selectInfluencersByIDs(ctx, influencerIDs)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	postResponses := buildPublicPostResponses(ctx, posts, images, tags, likedPostIDs, influencers)
	response.Write(w, response.BuildSuccess(postResponses, "", meta), http.StatusOK)
	return nil
}
