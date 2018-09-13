package jenkins_jr

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bukalapak/apinizer/response"
	"github.com/bukalapak/jenkins_jr/config"
	"github.com/bukalapak/jenkins_jr/pkg/api"
	"github.com/bukalapak/jenkins_jr/pkg/currentuser"
)

// HomepageResponse represents response to Homepage.
type HomepageResponse struct {
	Title    string                 `json:"title"`
	URL      string                 `json:"url"`
	Products []PostHomepageResponse `json:"products"`
}

// PostHomepageResponse represents posts of HomepageResponse.
type PostHomepageResponse struct {
	InspirationID int64  `json:"inspiration_id"`
	URL           string `json:"url"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
}

// Homepage sends as response posts to put in homepage.
func (env *Env) Homepage(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	form := env.buildHomepageSearchForm(r.URL.Query())
	posts, err := env.searchPost(ctx, form)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	user := currentuser.FromContext(ctx)
	hr := HomepageResponse{Title: config.InspirationHomepageTitle, URL: config.InspirationIndexURL(), Products: []PostHomepageResponse{}}
	m := api.IndexMeta{HTTPStatus: http.StatusOK, Limit: form.limit, Offset: form.offset, Total: len(posts)}

	if len(posts) == 0 || (user.Platform() == "blandroid" && user.AppVersion() <= 4025002) {
		response.Write(w, response.BuildSuccess(hr, "", m), http.StatusOK)
		return nil
	}

	env.buildHomepageProducts(ctx, posts, &hr)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	response.Write(w, response.BuildSuccess(hr, "", m), http.StatusOK)
	return nil
}

func (env *Env) buildHomepageProducts(ctx context.Context, posts []PostEntity, hr *HomepageResponse) error {
	postIDs := getPostIDsFromPosts(posts)
	images, err := env.findImagesByPostIDs(ctx, postIDs)
	if err != nil {
		return err
	}
	im := mapPostIDToImages(images)

	for _, post := range posts {
		imageURL := ""
		if len(im[post.ID]) >= 1 {
			image := im[post.ID][0]
			imageURL = image.styleURL(config.HomepageImageURLStyle)
		}

		hr.Products = append(hr.Products, PostHomepageResponse{
			InspirationID: post.ID,
			URL:           fmt.Sprintf("%s#%d", config.InspirationIndexURL(), post.ID),
			Description:   post.InfluencerName,
			ImageURL:      imageURL,
		})
	}
	return nil
}
