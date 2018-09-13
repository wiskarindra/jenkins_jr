package jenkins_jr

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bukalapak/apinizer/response"
	"github.com/wiskarindra/jenkins_jr/pkg/api"
	"github.com/wiskarindra/jenkins_jr/pkg/currentuser"
	"github.com/wiskarindra/jenkins_jr/pkg/log"
)

// PostLikeEntity represents post_likes entity.
type PostLikeEntity struct {
	Present bool `db:"present"` // is row present
	Liked   bool `db:"liked"`
}

// PostLikeResponse represents like in responses.
type PostLikeResponse struct {
	UserLike bool `json:"user_like"`
	Count    int  `json:"count"`
}

const (
	postLikeSelectQuery = "SELECT post_id FROM post_likes "
)

// Like likes/unlikes requested post.
func (env *Env) Like(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	postID, err := parseIntContext(ctx, api.GetParams(r).ByName("id"))
	if err != nil {
		response.Write(w, response.BuildError(api.InvalidParameterError), api.InvalidParameterError.HTTPCode)
		return api.InvalidParameterError
	}
	if !env.isPostPresent(ctx, postID) {
		response.Write(w, response.BuildError(api.PostNotFoundError), api.PostNotFoundError.HTTPCode)
		return api.PostNotFoundError
	}

	currentUser := currentuser.FromContext(ctx)
	like := env.findLike(ctx, postID, currentUser.ID)

	var query, operation, errorCategory string
	var factor int
	if !like.Present { // NEVER LIKED
		query = "INSERT INTO post_likes (post_id, bukalapak_user_id, liked) VALUES (?, ?, true)"
		operation = "like_count=like_count+1"
		errorCategory = "insert-sql-query"
		factor = 1
	} else if like.Liked { // WAS LIKED
		query = "UPDATE post_likes SET liked=false WHERE post_id=? AND bukalapak_user_id=?"
		operation = "like_count=like_count-1"
		errorCategory = "update-sql-query"
		factor = -1
	} else { // WAS UNLIKED
		query = "UPDATE post_likes SET liked=true WHERE post_id=? AND bukalapak_user_id=?"
		operation = "like_count=like_count+1"
		errorCategory = "update-sql-query"
		factor = 1
	}

	tx := env.DB.MustBegin()

	_, err = env.DB.ExecContext(ctx, query, postID, currentUser.ID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, errorCategory, "update/insert post_likes fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	var likeCount int
	err = tx.Get(&likeCount, "SELECT like_count FROM posts WHERE id=? FOR UPDATE", postID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "select-sql-query", "select like_count from posts fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}
	_, err = tx.Exec("UPDATE posts SET "+operation+" WHERE id=?", postID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "update-sql-query", "update posts like_count fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.ErrLog(ctx, err, "commit-sql-tx", "like post commit fail")
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	m := response.MetaInfo{HTTPStatus: http.StatusAccepted}
	response.Write(w, response.BuildSuccess(PostLikeResponse{UserLike: !like.Liked, Count: likeCount + factor}, "", m), http.StatusAccepted)
	return nil
}

func (env *Env) findLike(ctx context.Context, postID int64, userID uint) PostLikeEntity {
	query := "SELECT 1 AS present, liked FROM post_likes WHERE post_id=? AND bukalapak_user_id=? LIMIT 1"

	var like PostLikeEntity
	env.DB.GetContext(ctx, &like, query, postID, userID)
	return like
}

func (env *Env) findLikedPostIDs(ctx context.Context, postIDs []int64) ([]int64, error) {
	currentUser := currentuser.FromContext(ctx)
	if !isUserLoggedIn(currentUser.ID) {
		return nil, nil
	}
	condition := fmt.Sprintf("WHERE post_id IN (%s) AND bukalapak_user_id=%d AND liked=true", joinSliceInt64(postIDs), currentUser.ID)

	var likedPostIDs []int64
	err := env.DB.SelectContext(ctx, &likedPostIDs, postLikeSelectQuery+condition)
	if err != nil {
		log.ErrLog(ctx, err, "select-sql-query", "select from post_likes fail")
		return nil, err
	}
	return likedPostIDs, nil
}

func normalizeLikeCount(ctx context.Context, c int) int {
	if c < 0 {
		log.InfoLog(ctx, "negative like count", "bug", "negative-like-count")
		return 0
	}
	return c
}
