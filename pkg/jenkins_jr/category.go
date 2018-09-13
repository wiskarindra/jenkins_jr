package jenkins_jr

import (
	"context"
	"net/http"

	"github.com/bukalapak/apinizer/response"
	"github.com/wiskarindra/jenkins_jr/pkg/log"
	"github.com/jmoiron/sqlx"
)

// CategoryEntity represents categories entity.
type CategoryEntity struct {
	BukalapakCategoryID   int64  `db:"bukalapak_category_id"`
	BukalapakCategoryName string `db:"bukalapak_category_name"`
}

// CategoryResponse represents categories as response.
type CategoryResponse struct {
	BukalapakCategoryID   int64  `json:"id"`
	BukalapakCategoryName string `json:"name"`
}

const (
	categoryInsertQuery = "INSERT INTO categories (bukalapak_category_id, bukalapak_category_name, created_at, updated_at) VALUES (:bukalapak_category_id, :bukalapak_category_name, :created_at, :updated_at)"
	categorySelectQuery = "SELECT bukalapak_category_id, bukalapak_category_name FROM categories "
	categoryUpdateQuery = "UPDATE categories "
)

// Category sends as response categories of tags from published posts.
func (env *Env) Category(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	categories, err := env.findCategories(ctx)
	if err != nil {
		response.Write(w, response.BuildError(err), http.StatusUnprocessableEntity)
		return err
	}

	m := response.MetaInfo{HTTPStatus: http.StatusOK}
	response.Write(w, response.BuildSuccess(categories, "", m), http.StatusOK)
	return nil
}

func (env *Env) findCategories(ctx context.Context) ([]CategoryResponse, error) {
	query := categorySelectQuery + "WHERE count > 0 ORDER BY bukalapak_category_name"

	var categories []CategoryEntity
	if err := env.DB.SelectContext(ctx, &categories, query); err != nil {
		log.ErrLog(ctx, err, "select-sql-query", "select from categories fail")
		return nil, err
	}

	crs := []CategoryResponse{}
	for _, category := range categories {
		crs = append(crs, CategoryResponse{
			BukalapakCategoryID:   category.BukalapakCategoryID,
			BukalapakCategoryName: category.BukalapakCategoryName,
		})
	}
	return crs, nil
}

func addCategory(ctx context.Context, tx *sqlx.Tx, tag PostTagRequest) error {
	values := map[string]interface{}{
		"bukalapak_category_id":   tag.Category.ID,
		"bukalapak_category_name": tag.Category.Name,
		"created_at":              timeNowDB(),
		"updated_at":              timeNowDB(),
	}
	_, err := tx.NamedExecContext(ctx, categoryInsertQuery, values)
	return err
}

func decrementCategoryCounts(ctx context.Context, tx *sqlx.Tx, post PostEntity) error {
	for _, image := range post.Images {
		for _, tag := range image.Tags {
			err := decrementCategoryCount(ctx, tx, tag.BukalapakCategoryID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func decrementCategoryCount(ctx context.Context, tx *sqlx.Tx, bukalapakCategoryID int64) error {
	query := categoryUpdateQuery + "SET count=count-1, updated_at=? WHERE bukalapak_category_id=?"
	_, err := tx.ExecContext(ctx, query, timeNowDB(), bukalapakCategoryID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "update-sql-query", "update categories fail")
	}
	return err
}

func incrementCategoryCounts(ctx context.Context, tx *sqlx.Tx, post PostEntity) error {
	for _, image := range post.Images {
		for _, tag := range image.Tags {
			err := incrementCategoryCount(ctx, tx, tag.BukalapakCategoryID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func incrementCategoryCount(ctx context.Context, tx *sqlx.Tx, bukalapakCategoryID int64) error {
	query := categoryUpdateQuery + "SET count=count+1, updated_at=? WHERE bukalapak_category_id=?"
	_, err := tx.ExecContext(ctx, query, timeNowDB(), bukalapakCategoryID)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "update-sql-query", "update categories fail")
	}
	return err
}
