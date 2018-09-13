package jenkins_jr

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchPost(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	_, r := newRequestTest(nil, "jwt-valid")

	// Add published post
	env.DB.Exec("INSERT INTO posts (title, description, influencer_name, published ,created_at, updated_at) VALUES (?, ? , ?, ?, NOW(), NOW())", "Published Post", "Test Description", "Random Name", 1)
	// Add not published post
	env.DB.Exec("INSERT INTO posts (title, description, influencer_name, published ,created_at, updated_at) VALUES (?, ? , ?, ?, NOW(), NOW())", "Not Published Post", "Test Description", "Random Name", 0)

	// Search public post
	u, _ := url.Parse("http://example.com/posts?limit=10&offset=0")
	form := env.buildPublicPostSearchForm(u.Query())

	posts, err := env.searchPost(r.Context(), form)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(posts))

	// Search exclusive post
	u, _ = url.Parse("http://example.com/posts?limit=10&offset=0&keywords=Random Name")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	posts, err = env.searchPost(r.Context(), form)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(posts))

	// Search exclusive post with unknwon keyword
	u, _ = url.Parse("http://example.com/posts?limit=10&offset=0&keywords=Unknown Keyword")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	posts, err = env.searchPost(r.Context(), form)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(posts))
}

func TestSearchPostWithFilter(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	_, r := newRequestTest(nil, "jwt-valid")

	env.DB.Exec("INSERT INTO posts (title, description, influencer_name, published ,created_at, updated_at) VALUES ('Published Post', 'Test Description' , 'Random Name', '1', NOW(), NOW())")
	row, _ := env.DB.Exec("INSERT INTO posts (title, description, influencer_name, published ,created_at, updated_at) VALUES ('Published Post', 'Test Description' , 'Random Name', '1', NOW(), NOW())")

	postID, _ := row.LastInsertId()
	env.DB.Exec("INSERT INTO post_filters (post_id, bukalapak_category_id) VALUES (?, 144)", postID)

	// Search post without filter
	u, _ := url.Parse("http://example.com/posts?limit=10&offset=0")

	filter, ok := env.buildFilterForm(r.Context(), u.Query())
	assert.Equal(t, false, ok)

	form := env.buildPublicPostSearchForm(u.Query())
	assert.Equal(t, false, form.filtered)

	posts, err := env.searchPost(r.Context(), form)
	assert.Equal(t, nil, err)
	assert.Equal(t, 2, len(posts))
	assert.Equal(t, postID-1, posts[0].ID)
	assert.Equal(t, postID, posts[1].ID)

	// Search post with filter
	u, _ = url.Parse("http://example.com/posts?limit=10&offset=0&category_id=144")

	filter, ok = env.buildFilterForm(r.Context(), u.Query())
	assert.Equal(t, true, ok)

	filterEntities, err := env.filterPost(r.Context(), filter)
	assert.Equal(t, nil, err)

	form = env.buildPublicPostSearchForm(u.Query())
	form.applyFilters(filterEntities)
	assert.Equal(t, true, form.filtered)

	posts, err = env.searchPost(r.Context(), form)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(posts))
	assert.Equal(t, postID, posts[0].ID)
}

func TestBuildBasePostSearchForm(t *testing.T) {
	env := newEnvTest()

	// Normal case without sort
	u, _ := url.Parse("http://example.com/posts?limit=25&offset=100")
	form := env.buildBasePostSearchForm(u.Query())

	assert.Equal(t, uint64(25), form.limit)
	assert.Equal(t, uint64(100), form.offset)
	assert.Equal(t, "relevance", form.sort)

	assert.Equal(t, "", form.published)
	assert.Equal(t, "", form.keywords)
	assert.Equal(t, "", form.startDate)
	assert.Equal(t, "", form.finishDate)
	assert.Equal(t, "", form.dateRangeBy)

	// Limit and offset exceed allowed value
	u, _ = url.Parse("http://example.com/posts?limit=500&offset=1000000")
	form = env.buildBasePostSearchForm(u.Query())

	assert.Equal(t, uint64(10), form.limit)
	assert.Equal(t, uint64(0), form.offset)
	assert.Equal(t, "relevance", form.sort)

	// Limit and offset use negative value
	u, _ = url.Parse("http://example.com/posts?limit=-20&offset=-10")
	form = env.buildBasePostSearchForm(u.Query())

	assert.Equal(t, uint64(10), form.limit)
	assert.Equal(t, uint64(0), form.offset)
	assert.Equal(t, "relevance", form.sort)

	// Normal case with sort allowed value
	u, _ = url.Parse("http://example.com/posts?limit=20&offset=10&sort=popular")
	form = env.buildBasePostSearchForm(u.Query())

	assert.Equal(t, uint64(20), form.limit)
	assert.Equal(t, uint64(10), form.offset)
	assert.Equal(t, "popular", form.sort)

	// Normal case with sort not allowed value
	u, _ = url.Parse("http://example.com/posts?limit=20&offset=10&sort=bestselling")
	form = env.buildBasePostSearchForm(u.Query())

	assert.Equal(t, uint64(20), form.limit)
	assert.Equal(t, uint64(10), form.offset)
	assert.Equal(t, "relevance", form.sort)
}

func TestBuildPublicPostSearchForm(t *testing.T) {
	env := newEnvTest()

	// Normal case
	u, _ := url.Parse("http://example.com/posts?limit=25&offset=100")
	form := env.buildPublicPostSearchForm(u.Query())

	assert.Equal(t, uint64(25), form.limit)
	assert.Equal(t, uint64(100), form.offset)
	assert.Equal(t, "relevance", form.sort)

	assert.Equal(t, "true", form.published)
	assert.Equal(t, "", form.keywords)
	assert.Equal(t, "", form.startDate)
	assert.Equal(t, "", form.finishDate)
	assert.Equal(t, "", form.dateRangeBy)

	// Someone send published params
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&published=false")
	form = env.buildPublicPostSearchForm(u.Query())

	assert.Equal(t, form.published, "true")
}

func TestBuildExclusivePostSearchForm(t *testing.T) {
	env := newEnvTest()

	_, r := newRequestTest(nil, "jwt-valid")

	// Normal case
	u, _ := url.Parse("http://example.com/posts?limit=25&offset=100&published=false&keywords=Testing&date_range=2018-12-01:2018-12-03&date_range_by=created_at")
	form := env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, uint64(25), form.limit)
	assert.Equal(t, uint64(100), form.offset)
	assert.Equal(t, "created", form.sort)

	assert.Equal(t, "false", form.published)
	assert.Equal(t, "Testing", form.keywords)
	assert.Equal(t, "2018-11-30 17:00:00", form.startDate)
	assert.Equal(t, "2018-12-03 16:59:59", form.finishDate)
	assert.Equal(t, "created_at", form.dateRangeBy)

	// Normal case without published
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&keywords=Testing&date_range=2018-12-01:2018-12-03&date_range_by=created_at")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "", form.published)
	assert.Equal(t, "Testing", form.keywords)
	assert.Equal(t, "2018-11-30 17:00:00", form.startDate)
	assert.Equal(t, "2018-12-03 16:59:59", form.finishDate)
	assert.Equal(t, "created_at", form.dateRangeBy)

	// Not allowed date_range value
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&keywords=Testing&date_range=2018-12-01Z:2018-12-03&date_range_by=created_at")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "", form.published)
	assert.Equal(t, "Testing", form.keywords)
	assert.Equal(t, "", form.startDate)
	assert.Equal(t, "2018-12-03 16:59:59", form.finishDate)
	assert.Equal(t, "created_at", form.dateRangeBy)

	// Not allowed date_range value 2
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&keywords=Testing&date_range=2018-12-01:2018-12-03Z&date_range_by=created_at")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "", form.published)
	assert.Equal(t, "Testing", form.keywords)
	assert.Equal(t, "2018-11-30 17:00:00", form.startDate)
	assert.Equal(t, "", form.finishDate)
	assert.Equal(t, "created_at", form.dateRangeBy)

	// Normal case with half value date_range
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&keywords=Testing&date_range=2018-12-01:&date_range_by=created_at")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "Testing", form.keywords)
	assert.Equal(t, "2018-11-30 17:00:00", form.startDate)
	assert.Equal(t, "", form.finishDate)
	assert.Equal(t, "created_at", form.dateRangeBy)

	// Normal case with half value date_range 2
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&keywords=Testing&date_range=:2018-12-03&date_range_by=created_at")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "Testing", form.keywords)
	assert.Equal(t, "", form.startDate)
	assert.Equal(t, "2018-12-03 16:59:59", form.finishDate)
	assert.Equal(t, "created_at", form.dateRangeBy)

	// Normal case with date_range_by allowed value
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&date_range_by=published_at")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "", form.startDate)
	assert.Equal(t, "", form.finishDate)
	assert.Equal(t, "first_published_at", form.dateRangeBy)

	// Normal case with date_range_by allowed value 2
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100&date_range_by=modified_at")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "", form.startDate)
	assert.Equal(t, "", form.finishDate)
	assert.Equal(t, "updated_at", form.dateRangeBy)

	// Normal case without date_range_by
	u, _ = url.Parse("http://example.com/posts?limit=25&offset=100")
	form = env.buildExclusivePostSearchForm(r.Context(), u.Query())

	assert.Equal(t, "", form.published)
	assert.Equal(t, "", form.keywords)
	assert.Equal(t, "", form.startDate)
	assert.Equal(t, "", form.finishDate)
	assert.Equal(t, "", form.dateRangeBy)
}

func TestBuildHomepageSearchForm(t *testing.T) {
	env := newEnvTest()

	// Normal case
	u, _ := url.Parse("http://example.com/posts?limit=10&offset=0")
	form := env.buildHomepageSearchForm(u.Query())

	assert.Equal(t, uint64(10), form.limit)
	assert.Equal(t, uint64(0), form.offset)
	assert.Equal(t, "relevance", form.sort)
	assert.Equal(t, "true", form.published)

	// When offset > 0, it limit should set to 0
	u, _ = url.Parse("http://example.com/posts?limit=10&offset=10")
	form = env.buildHomepageSearchForm(u.Query())

	assert.Equal(t, uint64(0), form.limit)
	assert.Equal(t, uint64(10), form.offset)
	assert.Equal(t, "relevance", form.sort)
	assert.Equal(t, "true", form.published)
}

func TestBuildSelectedPostsSearchForm(t *testing.T) {
	env := newEnvTest()

	// Normal case
	u, _ := url.Parse("http://example.com/selected-posts")
	form := env.buildSelectedPostsSearchForm(u.Query())

	assert.Equal(t, uint64(3), form.limit)
	assert.Equal(t, uint64(0), form.offset)
	assert.Equal(t, "relevance", form.sort)
	assert.Equal(t, "true", form.published)

	// When limit is set
	limit := 4
	u, _ = url.Parse(fmt.Sprintf("http://example.com/selected-posts?limit=%d", limit))
	form = env.buildSelectedPostsSearchForm(u.Query())

	assert.Equal(t, uint64(limit), form.limit)
	assert.Equal(t, uint64(0), form.offset)
	assert.Equal(t, "relevance", form.sort)
	assert.Equal(t, "true", form.published)
}
