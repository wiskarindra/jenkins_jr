package jenkins_jr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/bukalapak/apinizer/response"
	"github.com/wiskarindra/jenkins_jr/config"
	"github.com/stretchr/testify/assert"
)

func TestHomepage(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	testHomepageEmpty(t, env)
	testHomepagePresent(t, env)
}

func testHomepageEmpty(t *testing.T, env Env) {
	w, r := newRequestTest(nil, "jwt-valid")

	err := env.Homepage(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data HomepageResponse  `json:"data"`
		Meta response.MetaInfo `json:"meta"`
	}
	json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, http.StatusOK, resp.Meta.HTTPStatus)
	assert.Equal(t, []PostHomepageResponse{}, resp.Data.Products)
}

func testHomepagePresent(t *testing.T, env Env) {
	w, r := newRequestTest(nil, "jwt-valid")

	query := "INSERT INTO posts (title, influencer_name, description, published, created_at, updated_at) VALUES (?, ? , ?, ?, NOW(), NOW())"
	influencerName := "Random Name"
	row, _ := env.DB.Exec(query, "Title 1", influencerName, "Random Description", true)

	query = "INSERT INTO post_images (post_id, url, height, width, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"
	postID, _ := row.LastInsertId()
	imageURL := "www.randomurl.com/randomimage.jpg"
	env.DB.Exec(query, postID, imageURL, 234, 435)

	err := env.Homepage(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data HomepageResponse  `json:"data"`
		Meta response.MetaInfo `json:"meta"`
	}
	json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, http.StatusOK, resp.Meta.HTTPStatus)
	assert.Equal(t, 1, len(resp.Data.Products))

	postURL := fmt.Sprintf("%s#%d", config.InspirationIndexURL(), postID)
	assert.Equal(t, postID, resp.Data.Products[0].InspirationID)
	assert.Equal(t, postURL, resp.Data.Products[0].URL)
	assert.Equal(t, influencerName, resp.Data.Products[0].Description)
	assert.Equal(t, imageURL, resp.Data.Products[0].ImageURL)
}
