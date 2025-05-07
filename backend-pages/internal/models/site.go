package models

import (
	"database/sql"
	"time"
)

type Site struct {
	ID        int            `json:"id"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	Domain    *string `json:"domain"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// MarshalJSON implements custom JSON marshaling for Site
// func (s Site) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(&struct {
// 		ID        int       `json:"id"`
// 		Code      string    `json:"code"`
// 		Name      string    `json:"name"`
// 		Domain    string    `json:"domain"`
// 		CreatedAt time.Time `json:"created_at"`
// 		UpdatedAt time.Time `json:"updated_at"`
// 	}{
// 		ID:        s.ID,
// 		Code:      s.Code,
// 		Name:      s.Name,
// 		Domain:    s.Domain.String,
// 		CreatedAt: s.CreatedAt,
// 		UpdatedAt: s.UpdatedAt,
// 	})
// }

type PageGroup struct {
	GroupID     int       `json:"group_id"`
	SiteID      int       `json:"site_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Menu        []Page    `json:"menu"`
}

type Page struct {
	ID           int       `json:"id"`
	SiteID       int       `json:"site_id"`
	GroupID      int       `json:"group_id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	ParentID     *int      `json:"parent_id"`
	Depth        int       `json:"depth"`
	MenuOrder    int       `json:"menu_order"`
	Content      sql.NullString    `json:"content"`
	IsPublished  bool      `json:"is_published"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Children     []Page    `json:"children,omitempty"`
}

type CreateSiteInput struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	
}

type CreatePageInput struct {
	SiteID    int     `json:"site_id"`
	GroupID   int     `json:"group_id"`
	Title     string  `json:"title"`
	Slug      string  `json:"slug"`
	ParentID  *int    `json:"parent_id"`
	Content   string  `json:"content"`
}

type CreatePageGroupInput struct {
	SiteID      int    `json:"site_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}