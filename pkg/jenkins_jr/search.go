package jenkins_jr

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bukalapak/jenkins_jr/config"
	"github.com/bukalapak/jenkins_jr/pkg/log"
)

// PostSearchForm represents search form.
type PostSearchForm struct {
	published string
	keywords  string

	startDate   string
	finishDate  string
	dateRangeBy string

	influencerID int64

	limit  uint64
	offset uint64
	sort   string

	filtered bool
	postIDs  []int64
}

var (
	availableSortOptions     = []string{"relevance", "date", "popular"}
	availablePublishedValues = []string{"1", "0", "true", "false"}
)

func (env *Env) searchPost(ctx context.Context, form PostSearchForm) ([]PostEntity, error) {
	var query bytes.Buffer
	query.WriteString(postSelectQuery)
	query.WriteString(form.buildWhereQuery())
	query.WriteString(form.buildOrderQuery())
	query.WriteString(form.buildLimitQuery())
	query.WriteString(form.buildOffsetQuery())

	log.DevLog(query.String(), form)

	var list []PostEntity
	err := env.DB.SelectContext(ctx, &list, query.String())
	if err != nil {
		log.ErrLog(ctx, err, "select-sql-query", "search posts fail")
		return nil, err
	}
	return list, nil
}

func (env *Env) buildBasePostSearchForm(params url.Values) PostSearchForm {
	limit, err := strconv.ParseUint(params.Get("limit"), 10, 8)
	if err != nil || limit > 25 || limit == 0 {
		limit = 10
	}

	offset, err := strconv.ParseUint(params.Get("offset"), 10, 8)
	if err != nil || offset > 10000 {
		offset = 0
	}

	sort := params.Get("sort")
	if !isInSliceString(sort, availableSortOptions) {
		sort = "relevance"
	}

	return PostSearchForm{limit: limit, offset: offset, sort: sort}
}

func (env *Env) buildPublicPostSearchForm(params url.Values) PostSearchForm {
	form := env.buildBasePostSearchForm(params)
	form.published = "true"
	return form
}

func (env *Env) buildExclusivePostSearchForm(ctx context.Context, params url.Values) PostSearchForm {
	// sort
	sort := "created"

	// published
	published := params.Get("published")
	if !isInSliceString(published, availablePublishedValues) {
		published = ""
	}

	// keywords
	reg, _ := regexp.Compile("[^a-zA-Z0-9 ]+")
	keywords := reg.ReplaceAllString(params.Get("keywords"), "")

	// startDate & finishDate
	dateRange := ":"
	if params.Get("date_range") != "" {
		dateRange = params.Get("date_range")
	}
	dateRanges := strings.Split(dateRange, ":")

	startDate := ""
	if dateRanges[0] != "" {
		st, err := time.Parse(config.DateRangeSearchFormat, fmt.Sprintf("%sT00:00:00+07:00", dateRanges[0]))
		if err != nil {
			log.ErrLog(ctx, err, "parse-params", "parse date_range fail")
		} else {
			startDate = st.UTC().Format(config.DatabaseDatetimeFormat)
		}
	}

	finishDate := ""
	if dateRanges[1] != "" {
		ft, err := time.Parse(config.DateRangeSearchFormat, fmt.Sprintf("%sT23:59:59+07:00", dateRanges[1]))
		if err != nil {
			log.ErrLog(ctx, err, "parse-params", "parse date_range fail")
		} else {
			finishDate = ft.UTC().Format(config.DatabaseDatetimeFormat)
		}
	}

	// dateRangeBy
	dateRangeBy := params.Get("date_range_by")
	switch dateRangeBy {
	case "published_at":
		dateRangeBy = "first_published_at"
	case "modified_at":
		dateRangeBy = "updated_at"
	case "created_at":
		dateRangeBy = "created_at"
	default:
		dateRangeBy = ""
	}

	form := env.buildBasePostSearchForm(params)
	form.sort = sort
	form.published = published
	form.keywords = keywords
	form.startDate = startDate
	form.finishDate = finishDate
	form.dateRangeBy = dateRangeBy
	return form
}

// form from buildHomepageSearchForm should made search to just return the first 10 post
// then return nothing when client request the next list
func (env *Env) buildHomepageSearchForm(params url.Values) PostSearchForm {
	form := env.buildBasePostSearchForm(params)
	form.sort = "relevance"
	form.published = "true"
	form.limit = 10
	if form.offset > 0 {
		form.limit = 0
	}

	return form
}

func (env *Env) buildSelectedPostsSearchForm(params url.Values) PostSearchForm {
	limit, err := strconv.ParseUint(params.Get("limit"), 10, 8)
	if err != nil || limit > 25 || limit == 0 {
		limit = 3
	}

	return PostSearchForm{
		limit:     limit,
		offset:    0,
		sort:      "relevance",
		published: "true",
	}
}

func (env *Env) buildInfluencerIndexPostSearchForm(params url.Values, influencerID int64) PostSearchForm {
	form := env.buildBasePostSearchForm(params)
	form.published = "true"
	form.influencerID = influencerID
	return form
}

func (env *Env) countPostTotalPage(ctx context.Context, form PostSearchForm) uint {
	query := "SELECT COUNT(1) FROM posts " + form.buildWhereQuery()

	var total uint
	env.DB.GetContext(ctx, &total, query)
	return uint(math.Ceil(float64(total) / float64(form.limit)))
}

func (form *PostSearchForm) buildWhereQuery() string {
	conditions := []string{"deleted=false"}
	if form.published != "" {
		conditions = append(conditions, fmt.Sprintf("published=%s", form.published))
	}
	if form.keywords != "" {
		conditions = append(conditions, "influencer_name LIKE '%"+form.keywords+"%'")
	}
	if form.dateRangeBy != "" {
		if form.startDate != "" {
			conditions = append(conditions, fmt.Sprintf("'%s' <= %s", form.startDate, form.dateRangeBy))
		}
		if form.finishDate != "" {
			conditions = append(conditions, fmt.Sprintf("%s <= '%s'", form.dateRangeBy, form.finishDate))
		}
	}

	if form.influencerID != 0 {
		conditions = append(conditions, fmt.Sprintf("influencer_id=%d", form.influencerID))
	}

	if form.filtered {
		pids := "NULL"
		if form.postIDs != nil {
			pids = joinSliceInt64(form.postIDs)
		}
		conditions = append(conditions, fmt.Sprintf("id IN (%s)", pids))
	}

	var query bytes.Buffer
	query.WriteString("WHERE ")
	query.WriteString(strings.Join(conditions, " AND "))
	query.WriteString(" ")
	return query.String()
}

func (form *PostSearchForm) buildOrderQuery() string {
	sort := ""
	switch form.sort {
	case "date":
		sort = "first_published_at DESC"
	case "popular":
		sort = "like_count DESC, first_published_at DESC"
	case "relevance":
		sort = "score DESC, first_published_at DESC"
	case "created":
		sort = "created_at DESC"
	}

	var query bytes.Buffer
	query.WriteString("ORDER BY ")
	query.WriteString(sort)
	query.WriteString(" ")
	return query.String()
}

func (form *PostSearchForm) buildLimitQuery() string {
	var query bytes.Buffer
	query.WriteString("LIMIT ")
	query.WriteString(strconv.FormatUint(form.limit, 10))
	query.WriteString(" ")
	return query.String()
}

func (form *PostSearchForm) buildOffsetQuery() string {
	var query bytes.Buffer
	query.WriteString("OFFSET ")
	query.WriteString(strconv.FormatUint(form.offset, 10))
	query.WriteString(" ")
	return query.String()
}
