package jenkins_jr

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bukalapak/apinizer/response"
	"github.com/stretchr/testify/assert"
)

func TestCategory(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	testCategoryEmpty(t, env)
	testCategoryCountZero(t, env)
	testCategoryCountMoreThanZero(t, env)
}

func testCategoryEmpty(t *testing.T, env Env) {
	w, r := newRequestTest(nil, "jwt-valid")

	err := env.Category(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []CategoryResponse `json:"data"`
		Meta response.MetaInfo  `json:"meta"`
	}
	json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, http.StatusOK, resp.Meta.HTTPStatus)
	assert.Equal(t, 0, len(resp.Data))
}

func testCategoryCountZero(t *testing.T, env Env) {
	w, r := newRequestTest(nil, "jwt-valid")

	_, err := env.DB.Exec("INSERT INTO categories (bukalapak_category_id, bukalapak_category_name, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", 32, "Random Category")
	assert.Equal(t, nil, err)

	err = env.Category(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []CategoryResponse `json:"data"`
		Meta response.MetaInfo  `json:"meta"`
	}
	err = json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, http.StatusOK, resp.Meta.HTTPStatus)
	assert.Equal(t, 0, len(resp.Data))
}

func testCategoryCountMoreThanZero(t *testing.T, env Env) {
	w, r := newRequestTest(nil, "jwt-valid")

	_, err := env.DB.Exec("INSERT INTO categories (bukalapak_category_id, bukalapak_category_name, count, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())", 144, "Random Category", 1)
	assert.Equal(t, nil, err)

	err = env.Category(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data []CategoryResponse `json:"data"`
		Meta response.MetaInfo  `json:"meta"`
	}
	err = json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, nil, err)

	assert.Equal(t, http.StatusOK, resp.Meta.HTTPStatus)
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, int64(144), resp.Data[0].BukalapakCategoryID)
	assert.Equal(t, "Random Category", resp.Data[0].BukalapakCategoryName)
}
