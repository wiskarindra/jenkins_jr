package jenkins_jr

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/wiskarindra/jenkins_jr/pkg/log"
	"github.com/jmoiron/sqlx"
)

// PostImageEntity represents post_images entity.
type PostImageEntity struct {
	ID       int64  `db:"id"`
	PostID   int64  `db:"post_id"`
	URL      string `db:"url"`
	Width    uint   `db:"width"`
	Height   uint   `db:"height"`
	Position int    `db:"position"`

	Tags []PostTagEntity
}

// PostImageRequest represents image in requests.
type PostImageRequest struct {
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	URL    string `json:"url"`

	Tags []PostTagRequest `json:"tags"`
}

// PostImageResponse represents image in responses.
type PostImageResponse struct {
	Width  uint   `json:"width"`
	Height uint   `json:"height"`
	URL    string `json:"url"`

	Tags []PostTagResponse `json:"tags"`
}

const (
	postImageInsertQuery = "INSERT INTO post_images (post_id, url, height, width, position, created_at, updated_at) VALUES (:post_id, :url, :height, :width, :position, :created_at, :updated_at)"
	postImageSelectQuery = "SELECT id, post_id, url, width, height FROM post_images "
)

func (env *Env) findImagesByPostID(ctx context.Context, postID int64) ([]PostImageEntity, error) {
	condition := fmt.Sprintf("WHERE post_id=%d ORDER BY position ASC", postID)
	return env.findImages(ctx, condition)
}

func (env *Env) findImagesByPostIDs(ctx context.Context, postIDs []int64) ([]PostImageEntity, error) {
	condition := fmt.Sprintf("WHERE post_id IN (%s) ORDER BY position ASC", joinSliceInt64(postIDs))
	return env.findImages(ctx, condition)
}

func (env *Env) findImages(ctx context.Context, condition string) ([]PostImageEntity, error) {
	query := postImageSelectQuery + condition

	var images []PostImageEntity
	err := env.DB.SelectContext(ctx, &images, query)
	if err != nil && err != sql.ErrNoRows {
		log.ErrLog(ctx, err, "select-sql-query", "select from post_images fail")
		return nil, err
	}
	return images, nil
}

func (ie PostImageEntity) styleURL(style string) string {
	return strings.Replace(ie.URL, "/original/", fmt.Sprintf("/%s/", style), 1)
}

func addImage(ctx context.Context, tx *sqlx.Tx, postID int64, image PostImageRequest, pos int) (sql.Result, error) {
	values := map[string]interface{}{
		"post_id":    postID,
		"url":        image.URL,
		"height":     image.Height,
		"width":      image.Width,
		"position":   pos,
		"created_at": timeNowDB(),
		"updated_at": timeNowDB(),
	}
	row, err := tx.NamedExecContext(ctx, postImageInsertQuery, values)
	if err != nil {
		tx.Rollback()
		log.ErrLog(ctx, err, "insert-sql-query", "insert to post_images fail")
	}
	return row, err
}

func buildImagesResponse(images []PostImageEntity) []PostImageResponse {
	irs := make([]PostImageResponse, 0)
	for _, image := range images {
		irs = append(irs, convertImageEntityToResponse(image))
	}
	return irs
}

func buildImagesResponseWithStyle(images []PostImageEntity, style string) []PostImageResponse {
	irs := make([]PostImageResponse, 0)
	for _, image := range images {
		irs = append(irs, convertImageEntityToResponseWithStyle(image, style))
	}
	return irs
}

func convertImageEntityToResponse(image PostImageEntity) PostImageResponse {
	return PostImageResponse{
		Width:  image.Width,
		Height: image.Height,
		URL:    image.URL,
		Tags:   buildTagsResponse(image.Tags),
	}
}

func convertImageEntityToResponseWithStyle(image PostImageEntity, style string) PostImageResponse {
	return PostImageResponse{
		Width:  image.Width,
		Height: image.Height,
		URL:    image.styleURL(style),
		Tags:   buildTagsResponse(image.Tags),
	}
}

func convertImageRequestToResponse(image PostImageRequest) PostImageResponse {
	return PostImageResponse{
		Width:  image.Width,
		Height: image.Height,
		URL:    image.URL,
		Tags:   convertTagsRequestToResponse(image.Tags),
	}
}

func convertImagesRequestToResponse(images []PostImageRequest) []PostImageResponse {
	resp := make([]PostImageResponse, 0)
	for _, image := range images {
		resp = append(resp, convertImageRequestToResponse(image))
	}
	return resp
}

func mapPostIDToImages(images []PostImageEntity) map[int64][]PostImageEntity {
	m := make(map[int64][]PostImageEntity)
	for _, image := range images {
		m[image.PostID] = append(m[image.PostID], image)
	}
	return m
}
