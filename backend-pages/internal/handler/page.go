package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"pages/internal/model"

	"pages/pkg/response"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) GetSiteMenu(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")

	fmt.Println(siteCode)

	rows, err := h.db.Query(`
        SELECT id, site_id, title, slug, parent_id, depth, 
               menu_order, content, is_published, created_at, updated_at
        FROM pages 
        WHERE site_id = ?
        ORDER BY depth, menu_order
    `, siteCode)
	if err != nil {
		response.Error(w, "Failed to fetch menu", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pages []*model.Page
	for rows.Next() {
		page := &model.Page{}
		err := rows.Scan(
			&page.ID,
			&page.SiteCode,
			&page.Title,
			&page.Slug,
			&page.ParentID,
			&page.Depth,
			&page.MenuOrder,
			&page.Content,
			&page.IsPublished,
			&page.CreatedAt,
			&page.UpdatedAt,
		)
		if err != nil {
			response.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		pages = append(pages, page)
		log.Println(page)
	}

	tree := buildMenuTree(pages)
	response.JSON(w, http.StatusOK, tree)
}

func (h *Handler) ListPages(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")

	rows, err := h.db.Query(`
        SELECT id, site_code, title, slug, parent_id, depth, 
               menu_order, content, is_published, created_at, updated_at
        FROM pages 
        WHERE site_code = ?
        ORDER BY depth, menu_order
    `, siteCode)
	if err != nil {
		response.Error(w, "Failed to fetch pages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pages []*model.Page
	for rows.Next() {
		page := &model.Page{}
		err := rows.Scan(
			&page.ID,
			&page.SiteCode,
			&page.Title,
			&page.Slug,
			&page.ParentID,
			&page.Depth,
			&page.MenuOrder,
			&page.Content,
			&page.IsPublished,
			&page.CreatedAt,
			&page.UpdatedAt,
		)
		if err != nil {
			response.Error(w, "Error scanning row", http.StatusInternalServerError)
			return
		}
		pages = append(pages, page)
	}

	response.JSON(w, http.StatusOK, pages)
}

func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string `json:"title"`
		Slug     string `json:"slug"`
		ParentID *uint  `json:"parent_id"`
		Content  string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	siteCode := chi.URLParam(r, "siteCode")

	tx, err := h.db.Begin()
	if err != nil {
		response.Error(w, "Transaction start failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var depth int
	if input.ParentID != nil {
		err := tx.QueryRow("SELECT depth FROM pages WHERE id = ?", *input.ParentID).Scan(&depth)
		if err != nil {
			response.Error(w, "Parent page not found", http.StatusBadRequest)
			return
		}
		depth++
	}

	result, err := tx.Exec(`
        INSERT INTO pages (
            site_code, title, slug, parent_id, depth, 
            menu_order, content, is_published, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, true, NOW(), NOW())
    `,
		siteCode,
		input.Title,
		input.Slug,
		input.ParentID,
		depth,
		0,
		input.Content,
	)
	if err != nil {
		response.Error(w, "Failed to create page", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()

	if err := tx.Commit(); err != nil {
		response.Error(w, "Transaction commit failed", http.StatusInternalServerError)
		return
	}

	var page model.Page
	err = h.db.QueryRow(`
        SELECT id, site_code, title, slug, parent_id, depth, 
               menu_order, content, is_published, created_at, updated_at
        FROM pages WHERE id = ?
    `, id).Scan(
		&page.ID,
		&page.SiteCode,
		&page.Title,
		&page.Slug,
		&page.ParentID,
		&page.Depth,
		&page.MenuOrder,
		&page.Content,
		&page.IsPublished,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err != nil {
		response.Error(w, "Failed to fetch created page", http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusCreated, page)
}

func (h *Handler) GetPage(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	var page model.Page
	err := h.db.QueryRow(`
        SELECT id, site_code, title, slug, parent_id, depth, 
               menu_order, content, is_published, created_at, updated_at
        FROM pages WHERE id = ?
    `, pageID).Scan(
		&page.ID,
		&page.SiteCode,
		&page.Title,
		&page.Slug,
		&page.ParentID,
		&page.Depth,
		&page.MenuOrder,
		&page.Content,
		&page.IsPublished,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		response.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	if err != nil {
		response.Error(w, "Failed to fetch page", http.StatusInternalServerError)
		return
	}

	response.JSON(w, http.StatusOK, page)
}

func (h *Handler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	var input struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(`
        UPDATE pages 
        SET title = ?, slug = ?, content = ?, updated_at = NOW()
        WHERE id = ?
    `,
		input.Title,
		input.Slug,
		input.Content,
		pageID,
	)
	if err != nil {
		response.Error(w, "Failed to update page", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		response.Error(w, "Failed to get rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		response.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Page updated successfully"})
}

func (h *Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	pageID := chi.URLParam(r, "pageID")

	result, err := h.db.Exec("DELETE FROM pages WHERE id = ?", pageID)
	if err != nil {
		response.Error(w, "Failed to delete page", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		response.Error(w, "Failed to get rows affected", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		response.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"message": "Page deleted successfully"})
}

func buildMenuTree(pages []*model.Page) []*model.PageTree {
	pageMap := make(map[uint]*model.PageTree)
	var roots []*model.PageTree

	for _, p := range pages {
		tree := &model.PageTree{
			Page:     p,
			Children: []*model.PageTree{},
		}
		pageMap[p.ID] = tree
	}

	for _, p := range pages {
		if !p.ParentID.Valid {
			roots = append(roots, pageMap[p.ID])
		} else {
			parentID := uint(p.ParentID.Int64)
			if parent, exists := pageMap[parentID]; exists {
				parent.Children = append(parent.Children, pageMap[p.ID])
			}
		}
	}

	return roots
}
