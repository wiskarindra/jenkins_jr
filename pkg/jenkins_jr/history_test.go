package jenkins_jr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/bukalapak/apinizer/response"
	"github.com/bukalapak/jenkins_jr/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	query := "INSERT INTO posts (title, influencer_name, description, published, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"
	row, _ := env.DB.Exec(query, "Title 1", "Random Name", "Random Description", true)
	postID, _ := row.LastInsertId()

	testHistoryEmpty(t, env, postID)
	testHistoryPresent(t, env, postID)
}

func testHistoryEmpty(t *testing.T, env Env, postID int64) {
	w, r := newRequestTest(nil, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err := env.History(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)
	if err != nil {
		return
	}

	var resp struct {
		Data []HistoryResponse `json:"data"`
		Meta response.MetaInfo `json:"meta"`
	}
	json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, http.StatusOK, resp.Meta.HTTPStatus)
	assert.Equal(t, 0, len(resp.Data))
}

func testHistoryPresent(t *testing.T, env Env, postID int64) {
	query := "INSERT INTO action_log_histories (record_id, record_type, changes, actor_id, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"
	env.DB.Exec(query, postID, "posts", `{"Title":["Then", "Now"]}`, 1)

	w, r := newRequestTest(nil, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err := env.History(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, http.StatusOK, w.Code)
	if err != nil {
		return
	}

	var resp struct {
		Data []HistoryResponse `json:"data"`
		Meta response.MetaInfo `json:"meta"`
	}
	json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, http.StatusOK, resp.Meta.HTTPStatus)
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "Then", resp.Data[0].Changes.Title[0])
	assert.Equal(t, "Now", resp.Data[0].Changes.Title[1])
}

func TestCountHistoryTotalPage(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	_, r := newRequestTest(nil, "jwt-valid")

	limit := 10
	for i := 0; i < limit; i++ {
		query := "INSERT INTO action_log_histories (record_id, record_type, actor_id, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())"
		env.DB.Exec(query, 1, "posts", 1)
	}
	assert.Equal(t, int(1), int(env.countHistoryTotalPage(r.Context(), 1, "posts", uint64(limit))))

	env.DB.Exec("INSERT INTO action_log_histories (record_id, record_type, actor_id, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())", 1, "posts", 1)
	assert.Equal(t, int(2), int(env.countHistoryTotalPage(r.Context(), 1, "posts", uint64(limit))))
}
