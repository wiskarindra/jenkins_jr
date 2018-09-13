package jenkins_jr

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"testing"

	"github.com/bukalapak/apinizer/response"
	"github.com/bukalapak/jenkins_jr/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestIsPostPresent(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	_, r := newRequestTest(nil, "jwt-valid")

	randPostID := rand.Int63()
	assert.Equal(t, false, env.isPostPresent(r.Context(), randPostID))

	row, _ := env.DB.Exec("INSERT INTO posts (title, influencer_name, description, created_at, updated_at) VALUES ('Title 1', 'Random Name', 'Random Description', NOW(), NOW())")
	postID, _ := row.LastInsertId()
	assert.Equal(t, true, env.isPostPresent(r.Context(), postID))
}

func TestFindPost(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	_, r := newRequestTest(nil, "jwt-valid")

	randPostID := rand.Int63()
	_, err := env.findPost(r.Context(), randPostID)
	assert.Equal(t, api.PostNotFoundError, err)

	// Insert normal post
	query := "INSERT INTO posts (title, influencer_name, description, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())"
	row, _ := env.DB.Exec(query, "Title 1", "Random Name", "Random Description")
	postID, _ := row.LastInsertId()
	post, err := env.findPost(r.Context(), postID)
	assert.Equal(t, nil, err)
	assert.Equal(t, postID, post.ID)

	// Insert deleted post
	row, _ = env.DB.Exec("INSERT INTO posts (title, influencer_name, description, deleted, created_at, updated_at) VALUES ('Title 1', 'Random Name', 'Random Description', true, NOW(), NOW())")
	postID, _ = row.LastInsertId()
	post, err = env.findPost(r.Context(), postID)
	assert.Equal(t, nil, err)
	assert.Equal(t, true, post.Deleted)
}

func TestIndex(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	// Add published post
	query := "INSERT INTO posts (title, description, influencer_name, published, first_published_at, last_published_at, created_at, updated_at) VALUES (?, ? , ?, ?, NOW(), NOW(), NOW(), NOW())"
	env.DB.Exec(query, "Published Post", "Test Description", "Random Name", true)
	// Add not published post
	query = "INSERT INTO posts (title, description, influencer_name, published, first_published_at, last_published_at, created_at, updated_at) VALUES (?, ? , ?, ?, NOW(), NOW(), NOW(), NOW())"
	env.DB.Exec(query, "Not Published Post", "Test Description", "Random Name", false)

	w, r := newRequestTest(nil, "jwt-valid")

	err := env.PublicIndex(w, r)
	assert.Equal(t, nil, err)

	var resp struct {
		Data []PostPublicResponse `json:"data"`
		Meta response.MetaInfo    `json:"meta"`
	}

	json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, 200, resp.Meta.HTTPStatus)
	assert.Equal(t, 1, len(resp.Data))
	assert.Equal(t, "Published Post", resp.Data[0].Title)
	assert.Equal(t, "Test Description", resp.Data[0].Description)
	assert.Equal(t, "Random Name", resp.Data[0].InfluencerName)
}

func TestCreate(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	reqBody, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/bukalapak/jenkins_jr/testdata/request/post.json")
	assert.Equal(t, nil, err)
	w, r := newRequestTest(reqBody, "jwt-valid")

	err = env.Create(w, r)
	assert.Equal(t, nil, err)

	var pr PostRequest
	json.Unmarshal([]byte(reqBody), &pr)

	var images []PostImageEntity
	env.DB.Select(&images, postImageSelectQuery)

	assert.Equal(t, len(pr.Images), len(images))
	for i, image := range images {
		assert.Equal(t, pr.Images[i].Width, image.Width)
		assert.Equal(t, pr.Images[i].Height, image.Height)
		assert.Equal(t, pr.Images[i].URL, image.URL)
		assert.Equal(t, i, image.Position)

		var tags []PostTagEntity
		env.DB.Select(&tags, postTagSelectQuery)

		assert.Equal(t, len(pr.Images[i].Tags), len(tags))
		for j, tag := range image.Tags {
			assert.Equal(t, pr.Images[i].Tags[j].URL, tag.URL)
			assert.Equal(t, pr.Images[i].Tags[j].Name, tag.Name)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.X, tag.CoordX)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.Y, tag.CoordY)
			assert.Equal(t, pr.Images[i].Tags[j].Category.ID, tag.BukalapakCategoryID)
		}
	}

	var resp struct {
		Data PostExclusiveResponse `json:"data"`
		Meta response.MetaInfo     `json:"meta"`
	}
	json.Unmarshal([]byte(w.Body.String()), &resp)
	assert.Equal(t, 201, resp.Meta.HTTPStatus)
	assert.Equal(t, pr.Title, resp.Data.Title)
	assert.Equal(t, pr.InfluencerName, resp.Data.InfluencerName)
	assert.Equal(t, pr.Published, resp.Data.Published)

	assert.Equal(t, len(pr.Images), len(resp.Data.Images))
	for i, image := range resp.Data.Images {
		assert.Equal(t, pr.Images[i].Width, image.Width)
		assert.Equal(t, pr.Images[i].Height, image.Height)
		assert.Equal(t, pr.Images[i].URL, image.URL)

		assert.Equal(t, len(pr.Images[i].Tags), len(image.Tags))
		for j, tag := range image.Tags {
			assert.Equal(t, pr.Images[i].Tags[j].Name, tag.Name)
			assert.Equal(t, pr.Images[i].Tags[j].URL, tag.URL)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.X, tag.Coord.X)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.Y, tag.Coord.Y)
		}
	}

	// Category-related
	var categories []CategoryEntity
	env.DB.Select(&categories, categorySelectQuery)
	assert.Equal(t, 1, len(categories))
	assert.Equal(t, int64(32), categories[0].BukalapakCategoryID)
	assert.Equal(t, "Category name 1", categories[0].BukalapakCategoryName)

	var filters []PostFilterEntity
	env.DB.Select(&filters, postFilterSelectQuery)
	assert.Equal(t, 1, len(filters))
	assert.Equal(t, int64(32), filters[0].BukalapakCategoryID)
}

func TestUpdate(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	row, _ := env.DB.Exec("INSERT INTO posts (title, description, influencer_name, published, like_count, created_at, updated_at) VALUES ('Published Post', 'Test Description' , 'Tester', 1, 10, NOW(), NOW())")
	postID, _ := row.LastInsertId()

	reqBody, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/bukalapak/jenkins_jr/testdata/request/post.json")
	assert.Equal(t, nil, err)
	w, r := newRequestTest(reqBody, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err = env.Update(w, r)
	assert.Equal(t, nil, err)

	var pr PostRequest
	json.Unmarshal([]byte(reqBody), &pr)

	var images []PostImageEntity
	env.DB.Select(&images, postImageSelectQuery)

	assert.Equal(t, len(pr.Images), len(images))
	for i, image := range images {
		assert.Equal(t, pr.Images[i].Width, image.Width)
		assert.Equal(t, pr.Images[i].Height, image.Height)
		assert.Equal(t, pr.Images[i].URL, image.URL)
		assert.Equal(t, i, image.Position)

		var tags []PostTagEntity
		env.DB.Select(&tags, postTagSelectQuery)

		assert.Equal(t, len(pr.Images[i].Tags), len(tags))
		for j, tag := range image.Tags {
			assert.Equal(t, pr.Images[i].Tags[j].URL, tag.URL)
			assert.Equal(t, pr.Images[i].Tags[j].Name, tag.Name)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.X, tag.CoordX)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.Y, tag.CoordY)
			assert.Equal(t, pr.Images[i].Tags[j].Category.ID, tag.BukalapakCategoryID)
		}
	}

	var resp struct {
		Data PostExclusiveResponse `json:"data"`
		Meta response.MetaInfo     `json:"meta"`
	}
	json.Unmarshal([]byte(w.Body.String()), &resp)

	assert.Equal(t, 202, resp.Meta.HTTPStatus)
	assert.Equal(t, pr.Title, resp.Data.Title)
	assert.Equal(t, pr.InfluencerName, resp.Data.InfluencerName)
	assert.Equal(t, pr.Published, resp.Data.Published)

	assert.Equal(t, len(pr.Images), len(resp.Data.Images))
	for i, image := range resp.Data.Images {
		assert.Equal(t, pr.Images[i].Width, image.Width)
		assert.Equal(t, pr.Images[i].Height, image.Height)
		assert.Equal(t, pr.Images[i].URL, image.URL)

		assert.Equal(t, len(pr.Images[i].Tags), len(image.Tags))
		for j, tag := range image.Tags {
			assert.Equal(t, pr.Images[i].Tags[j].Name, tag.Name)
			assert.Equal(t, pr.Images[i].Tags[j].URL, tag.URL)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.X, tag.Coord.X)
			assert.Equal(t, pr.Images[i].Tags[j].Coord.Y, tag.Coord.Y)
		}
	}
}

func TestUpdateCategories(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	query := "INSERT INTO posts (title, description, influencer_name, published, like_count, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())"
	row, _ := env.DB.Exec(query, "Published Post", "Test Description", "Tester", 1, 10)
	postID, _ := row.LastInsertId()

	reqBody, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/bukalapak/jenkins_jr/testdata/request/post.json")
	assert.Equal(t, nil, err)
	w, r := newRequestTest(reqBody, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err = env.Update(w, r)
	assert.Equal(t, nil, err)

	var categories []CategoryEntity
	err = env.DB.Select(&categories, categorySelectQuery+"WHERE count > 0")
	assert.Equal(t, 1, len(categories))
	assert.Equal(t, int64(32), categories[0].BukalapakCategoryID)

	var filters []PostFilterEntity
	err = env.DB.Select(&filters, postFilterSelectQuery+"WHERE bukalapak_category_id=32")
	assert.Equal(t, 1, len(filters))
	assert.Equal(t, postID, filters[0].PostID)
	assert.Equal(t, int64(32), filters[0].BukalapakCategoryID)

	// Another request
	reqBody, err = ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/bukalapak/jenkins_jr/testdata/request/post2.json")
	assert.Equal(t, nil, err)
	w, r = newRequestTest(reqBody, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err = env.Update(w, r)
	assert.Equal(t, nil, err)

	categories = nil
	err = env.DB.Select(&categories, categorySelectQuery+"WHERE count > 0")
	assert.Equal(t, 2, len(categories))
	assert.Equal(t, int64(77), categories[0].BukalapakCategoryID)
	assert.Equal(t, int64(144), categories[1].BukalapakCategoryID)

	filters = nil
	err = env.DB.Select(&filters, postFilterSelectQuery+"WHERE bukalapak_category_id=77")
	assert.Equal(t, 1, len(filters))
	assert.Equal(t, postID, filters[0].PostID)
	assert.Equal(t, int64(77), filters[0].BukalapakCategoryID)
}

func TestDelete(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	// Before Deletion
	reqBody, err := ioutil.ReadFile(os.Getenv("GOPATH") + "/src/github.com/bukalapak/jenkins_jr/testdata/request/post.json")
	assert.Equal(t, nil, err)

	w, r := newRequestTest(reqBody, "jwt-valid")

	err = env.Create(w, r)
	assert.Equal(t, nil, err)

	var posts []PostEntity
	env.DB.Select(&posts, postSelectQuery+"WHERE deleted=false")
	assert.Equal(t, 1, len(posts))
	postID := posts[0].ID

	// Categories
	var categories []CategoryEntity
	env.DB.Select(&categories, categorySelectQuery+"WHERE count > 0")
	assert.Equal(t, 1, len(categories))

	// Filters
	var filters []PostFilterEntity
	env.DB.Select(&filters, postFilterSelectQuery)
	assert.Equal(t, 1, len(filters))

	// After Deletion
	w, r = newRequestTest(nil, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err = env.Delete(w, r)
	assert.Equal(t, nil, err)

	// Posts
	posts = nil
	env.DB.Select(&posts, postSelectQuery+"WHERE deleted=false")
	assert.Equal(t, 0, len(posts))

	// Categories
	categories = nil
	env.DB.Select(&categories, categorySelectQuery+"WHERE count > 0")
	assert.Equal(t, 0, len(categories))

	// Filters
	filters = nil
	env.DB.Select(&filters, postFilterSelectQuery)
	assert.Equal(t, 0, len(filters))
}

func TestLike(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	// Add published post with 10 like_count
	query := "INSERT INTO posts (title, description, influencer_name, published, like_count, created_at, updated_at) VALUES (?, ? , ?, ?, ?, NOW(), NOW())"
	row, _ := env.DB.Exec(query, "Published Post", "Test Description", "Tester", 1, 10)
	postID, _ := row.LastInsertId()

	w, r := newRequestTest(nil, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err := env.Like(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"data":{"user_like":true,"count":11},"meta":{"http_status":202}}`+"\n", w.Body.String())

	var le struct {
		PostID int64 `db:"post_id"`
		Liked  bool  `db:"liked"`
	}
	env.DB.Get(&le, "SELECT post_id, liked FROM post_likes ORDER BY id DESC LIMIT 1")
	assert.Equal(t, postID, le.PostID)
	assert.Equal(t, true, le.Liked)

	post, _ := env.findPost(r.Context(), postID)
	assert.Equal(t, 11, post.LikeCount)

	w, r = newRequestTest(nil, "jwt-valid")
	r = api.SetParams(r, map[string]string{"id": fmt.Sprintf("%d", postID)})

	err = env.Like(w, r)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"data":{"user_like":false,"count":10},"meta":{"http_status":202}}`+"\n", w.Body.String())

	env.DB.Get(&le, "SELECT post_id, liked FROM post_likes ORDER BY id DESC LIMIT 1")
	assert.Equal(t, postID, le.PostID)
	assert.Equal(t, false, le.Liked)

	post, _ = env.findPost(r.Context(), postID)
	assert.Equal(t, 10, post.LikeCount)
}

func TestCountPostTotalPage(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	_, r := newRequestTest(nil, "jwt-valid")

	u, _ := url.Parse("http://example.com/posts?limit=10&offset=0")
	form := env.buildPublicPostSearchForm(u.Query())

	limit := 10
	for i := 0; i < limit; i++ {
		query := "INSERT INTO posts (title, description, influencer_name, published, like_count, created_at, updated_at) VALUES (?, ? , ?, ?, ?, NOW(), NOW())"
		env.DB.Exec(query, "Published Post", "Test Description", "Tester", 1, 10)
	}
	assert.Equal(t, int(1), int(env.countPostTotalPage(r.Context(), form)))

	query := "INSERT INTO posts (title, description, influencer_name, published, like_count, created_at, updated_at) VALUES (?, ? , ?, ?, ?, NOW(), NOW())"
	env.DB.Exec(query, "Published Post", "Test Description", "Tester", 1, 10)
	assert.Equal(t, int(2), int(env.countPostTotalPage(r.Context(), form)))
}
