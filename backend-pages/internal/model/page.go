package model

import (
	"database/sql"
	"time"
)

type Page struct {
	ID          uint
	SiteCode    string
	Title       string
	Slug        string
	ParentID    sql.NullInt64
	Depth       int
	MenuOrder   int
	Content     sql.NullString
	IsPublished bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PageTree struct {
	*Page
	Children []*PageTree `json:"children"`
}
