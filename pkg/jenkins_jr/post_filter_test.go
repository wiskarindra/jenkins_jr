package jenkins_jr

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterPost(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	query := "INSERT INTO post_filters (post_id, bukalapak_category_id) VALUES (?, ?)"
	env.DB.Exec(query, 1, 144)
	env.DB.Exec(query, 1, 27)
	env.DB.Exec(query, 2, 27)
	env.DB.Exec(query, 3, 27)

	testFilterOneResult(t, env)
	testFilterMoreThanOneResult(t, env)
}

func testFilterOneResult(t *testing.T, env Env) {
	_, r := newRequestTest(nil, "jwt-valid")

	u, _ := url.Parse("http://example.com/posts?category_id=144")
	form, ok := env.buildFilterForm(r.Context(), u.Query())

	assert.Equal(t, int64(144), form.bukalapakCategoryID)
	assert.Equal(t, true, ok)

	filters, err := env.filterPost(r.Context(), form)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(filters))
}

func testFilterMoreThanOneResult(t *testing.T, env Env) {
	_, r := newRequestTest(nil, "jwt-valid")

	u, _ := url.Parse("http://example.com/posts?category_id=27")
	form, ok := env.buildFilterForm(r.Context(), u.Query())

	assert.Equal(t, int64(27), form.bukalapakCategoryID)
	assert.Equal(t, true, ok)

	filters, err := env.filterPost(r.Context(), form)
	assert.Equal(t, nil, err)
	assert.Equal(t, 3, len(filters))
	assert.Equal(t, int64(1), filters[0].PostID)
	assert.Equal(t, int64(2), filters[1].PostID)
	assert.Equal(t, int64(3), filters[2].PostID)
}
