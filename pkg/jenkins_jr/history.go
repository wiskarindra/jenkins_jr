package jenkins_jr

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/bukalapak/apinizer/response"
	"github.com/bukalapak/jenkins_jr/config"
	"github.com/bukalapak/jenkins_jr/pkg/api"
	"github.com/bukalapak/jenkins_jr/pkg/currentuser"
	"github.com/bukalapak/jenkins_jr/pkg/log"
)

// HistoryEntity represents histories entity.
type HistoryEntity struct {
	Changes   string     `db:"changes"`
	ActorID   int64      `db:"actor_id"`
	CreatedAt *time.Time `db:"created_at"`
}

// HistoryResponse represents histories as response.
type HistoryResponse struct {
	Changes   *PostHistoryChanges `json:"changes"`
	ActorID   int64               `json:"actor_id"`
	CreatedAt string              `json:"created_at"`
}

// PostHistoryData represents changes in a post.
type PostHistoryData struct {
	PostID  int64
	ActorID int64
	Was     PostEntity
	Now     PostEntity
}

// PostHistoryChanges represents changes in post fields.
type PostHistoryChanges struct {
	Title          []string `json:"title,omitempty"`
	Description    []string `json:"description,omitempty"`
	InfluencerName []string `json:"influencer_name,omitempty"`
	Published      []string `json:"published,omitempty"`

	FirstPublishedAt []string `json:"first_published_at,omitempty"`
	LastPublishedAt  []string `json:"last_published_at,omitempty"`
	Deleted          []string `json:"deleted,omitempty"`
	CreatedAt        []string `json:"created_at,omitempty"`
	UpdatedAt        []string `json:"updated_at,omitempty"`

	Images [][]PostImageEntity `json:"images,omitempty"`
}

const (
	historyInsertQuery = "INSERT INTO action_log_histories (record_id, record_type, changes, actor_id, created_at, updated_at) VALUES (:record_id, :record_type, :changes, :actor_id, :created_at, :created_at)"
	historySelectQuery = "SELECT changes, actor_id, created_at FROM action_log_histories "
	historyCountQuery  = "SELECT COUNT(1) FROM action_log_histories "
)

// History sends as response histories of requested post.
func (env *Env) History(w http.ResponseWriter, r *http.Request) error {
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

	limit, offset := getLimitOffsetFromURLQuery(r.URL.Query())
	histories, err := env.findHistories(ctx, postID, limit, offset)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	totalPages := env.countHistoryTotalPage(ctx, postID, "posts", limit)
	m := api.IndexMeta{HTTPStatus: http.StatusOK, Limit: limit, Offset: offset, Total: len(histories), TotalPages: totalPages}
	response.Write(w, response.BuildSuccess(histories, "", m), http.StatusOK)
	return nil
}

func (env *Env) addPostLogHistory(ctx context.Context, phd PostHistoryData) {
	currentUser := currentuser.FromContext(ctx)
	phd.ActorID = int64(currentUser.ID)

	post, err := env.findPost(ctx, phd.PostID)
	if err != nil {
		return
	}
	phd.Now = post

	changes, err := phd.buildChanges()
	if err != nil {
		log.ErrLog(ctx, err, "marshal-to-json", "marshal changes to json fail")
		return
	}
	if changes == "{}" {
		return
	}

	values := map[string]interface{}{
		"record_id":   phd.PostID,
		"record_type": "posts",
		"changes":     changes,
		"actor_id":    phd.ActorID,
		"created_at":  timeNowDB(),
	}
	_, err = env.DB.NamedExecContext(ctx, historyInsertQuery, values)
	if err != nil {
		log.ErrLog(ctx, err, "insert-sql-query", "insert to action_log_histories fail")
	}
}

func (env *Env) countHistoryTotalPage(ctx context.Context, rID int64, rType string, limit uint64) uint {
	query := historyCountQuery + "WHERE record_id=? AND record_type=?"

	var total uint
	env.DB.GetContext(ctx, &total, query, rID, rType)
	return uint(math.Ceil(float64(total) / float64(limit)))
}

func (env *Env) findHistories(ctx context.Context, postID int64, limit uint64, offset uint64) ([]HistoryResponse, error) {
	query := historySelectQuery + "WHERE record_id=? AND record_type='posts' ORDER BY created_at DESC LIMIT ? OFFSET ?"

	var histories []HistoryEntity
	err := env.DB.SelectContext(ctx, &histories, query, postID, limit, offset)
	if err != nil {
		log.ErrLog(ctx, err, "select-sql-query", "select from action_log_histories fail")
		return nil, err
	}

	hrs := []HistoryResponse{}
	for _, history := range histories {
		changes := &PostHistoryChanges{}
		json.Unmarshal([]byte(history.Changes), changes)
		hrs = append(hrs, HistoryResponse{
			Changes:   changes,
			ActorID:   history.ActorID,
			CreatedAt: history.CreatedAt.Format(config.ResponseDatetimeFormat),
		})
	}
	return hrs, nil
}

func (phd *PostHistoryData) buildChanges() (string, error) {
	raw := map[string][2]string{
		"title":           [2]string{phd.Was.Title, phd.Now.Title},
		"description":     [2]string{phd.Was.Description, phd.Now.Description},
		"influencer_name": [2]string{phd.Was.InfluencerName, phd.Now.InfluencerName},
		"published":       [2]string{fmt.Sprintf("%v", phd.Was.Published), fmt.Sprintf("%v", phd.Now.Published)},
		"deleted":         [2]string{fmt.Sprintf("%v", phd.Was.Deleted), fmt.Sprintf("%v", phd.Now.Deleted)},

		"first_published_at": [2]string{timeToStringResponse(phd.Was.FirstPublishedAt), timeToStringResponse(phd.Now.FirstPublishedAt)},
		"last_published_at":  [2]string{timeToStringResponse(phd.Was.LastPublishedAt), timeToStringResponse(phd.Now.LastPublishedAt)},
		"created_at":         [2]string{timeToStringResponse(phd.Was.CreatedAt), timeToStringResponse(phd.Now.CreatedAt)},
		"updated_at":         [2]string{timeToStringResponse(phd.Was.UpdatedAt), timeToStringResponse(phd.Now.UpdatedAt)},
	}

	changes := make(map[string]interface{})
	for k, v := range raw {
		if v[0] != v[1] {
			changes[k] = [2]string{v[0], v[1]}
		}
	}

	// TODO: handle images history better
	// changes["images"] = [][]PostImageEntity{phd.Was.Images, phd.Now.Images}

	json, err := json.Marshal(changes)
	return string(json), err
}
