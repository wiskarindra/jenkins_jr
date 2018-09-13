package jenkins_jr

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bukalapak/jenkins_jr/pkg/log"
)

type PostFilterForm struct {
	bukalapakCategoryID int64
}

type PostFilterEntity struct {
	PostID              int64 `db:"post_id"`
	BukalapakCategoryID int64 `db:"bukalapak_category_id"`
}

const (
	postFilterSelectQuery = "SELECT post_id, bukalapak_category_id FROM post_filters "
)

func (env *Env) filterPost(ctx context.Context, form PostFilterForm) ([]PostFilterEntity, error) {
	var query bytes.Buffer
	query.WriteString(postFilterSelectQuery)
	query.WriteString(form.buildFilterQuery())

	log.DevLog(query.String())
	log.DevLog(form)

	var filters []PostFilterEntity
	err := env.DB.SelectContext(ctx, &filters, query.String())
	if err != nil {
		log.ErrLog(ctx, err, "select-sql-query", "filter posts fail")
		return nil, err
	}

	return filters, nil
}

func (env *Env) buildFilterForm(ctx context.Context, params url.Values) (PostFilterForm, bool) {
	form := PostFilterForm{}
	ok := false

	categoryID := params.Get("category_id")
	if categoryID != "" {
		categoryIDParsed, err := strconv.ParseInt(categoryID, 10, 64)
		if err != nil {
			log.ErrLog(ctx, err, "parse-params", "parse postID fail")
		} else {
			form.bukalapakCategoryID = categoryIDParsed
			ok = true
		}
	}

	return form, ok
}

func (form *PostSearchForm) applyFilters(filters []PostFilterEntity) {
	form.filtered = true
	for _, f := range filters {
		form.postIDs = append(form.postIDs, f.PostID)
	}
}

func (form PostFilterForm) buildFilterQuery() string {
	conditions := []string{}
	conditions = append(conditions, fmt.Sprintf("bukalapak_category_id=%d", form.bukalapakCategoryID))

	var query bytes.Buffer
	query.WriteString("WHERE ")
	query.WriteString(strings.Join(conditions, " AND "))
	query.WriteString(" ")
	return query.String()
}
