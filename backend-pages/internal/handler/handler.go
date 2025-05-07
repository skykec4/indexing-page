package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"pages/internal/models"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) GetSites(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT site_id, code, name, domain, created_at, updated_at FROM sites")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var site models.Site
		if err := rows.Scan(&site.ID, &site.Code, &site.Name, &site.Domain, &site.CreatedAt, &site.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sites = append(sites, site)
	}

	json.NewEncoder(w).Encode(sites)
}

func (h *Handler) CreateSite(w http.ResponseWriter, r *http.Request) {
	var input models.CreateSiteInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(
		"INSERT INTO sites (code, name, domain) VALUES (?, ?, ?)",
		input.Code, input.Name, input.Domain,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"created": true,
		"site_id": id,
	})
}

func (h *Handler) GetSiteMenu(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")
	fmt.Println("siteCode : ",siteCode)

	// 사이트 정보 조회
	var site models.Site
	err := h.db.QueryRow(
		"SELECT site_id, code, name, domain, created_at, updated_at FROM sites WHERE code = ?",
		siteCode,
	).Scan(&site.ID, &site.Code, &site.Name, &site.Domain, &site.CreatedAt, &site.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "사이트를 찾을 수 없습니다", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 그룹 조회
	groupRows, err := h.db.Query(
		"SELECT group_id, site_id, name, description, created_at, updated_at FROM page_groups WHERE site_id = ? ORDER BY name",
		site.ID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer groupRows.Close()

	var pageGroups []models.PageGroup
	for groupRows.Next() {
		var group models.PageGroup
		if err := groupRows.Scan(&group.GroupID, &group.SiteID, &group.Name, &group.Description, &group.CreatedAt, &group.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// 각 그룹의 메뉴 조회
		pageRows, err := h.db.Query(
			`SELECT id, site_id, group_id, title, slug, parent_id, depth, menu_order, 
			content, is_published, created_at, updated_at 
			FROM pages 
			WHERE site_id = ? AND group_id = ? 
			ORDER BY depth, menu_order`,
			site.ID, group.GroupID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer pageRows.Close()

		var pages []models.Page
		for pageRows.Next() {
			var page models.Page
			if err := pageRows.Scan(
				&page.ID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
				&page.ParentID, &page.Depth, &page.MenuOrder, &page.Content,
				&page.IsPublished, &page.CreatedAt, &page.UpdatedAt,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			pages = append(pages, page)
		}

		group.Menu = buildMenuTree(pages)
		pageGroups = append(pageGroups, group)
	}

	response := struct {
		models.Site
		PageGroups []models.PageGroup `json:"page_groups"`
	}{
		Site:       site,
		PageGroups: pageGroups,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	var input models.CreatePageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 부모 페이지가 있는 경우 depth 계산
	depth := 0
	if input.ParentID != nil {
		err := h.db.QueryRow("SELECT depth FROM pages WHERE id = ?", *input.ParentID).Scan(&depth)
		if err != nil {
			http.Error(w, "부모 페이지를 찾을 수 없습니다", http.StatusBadRequest)
			return
		}
		depth++
	}

	result, err := h.db.Exec(
		`INSERT INTO pages (site_id, group_id, title, slug, parent_id, depth, menu_order, content, is_published) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.SiteID, input.GroupID, input.Title, input.Slug, input.ParentID, depth, 0, input.Content, true,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 생성된 페이지 조회
	var page models.Page
	err = h.db.QueryRow(
		"SELECT * FROM pages WHERE id = ?", id,
	).Scan(
		&page.ID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
		&page.ParentID, &page.Depth, &page.MenuOrder, &page.Content,
		&page.IsPublished, &page.CreatedAt, &page.UpdatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(page)
}

func buildMenuTree(pages []models.Page) []models.Page {
	pageMap := make(map[int]*models.Page)
	var roots []models.Page

	// 모든 페이지를 맵에 저장
	for i := range pages {
		pageMap[pages[i].ID] = &pages[i]
	}

	// 트리 구조 구성
	for _, page := range pages {
		if page.ParentID == nil {
			roots = append(roots, page)
		} else {
			if parent, exists := pageMap[*page.ParentID]; exists {
				parent.Children = append(parent.Children, page)
			}
		}
	}

	return roots
}

func (h *Handler) ListPages(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")

	rows, err := h.db.Query(`
		SELECT id, site_id, group_id, title, slug, parent_id, depth, 
		menu_order, content, is_published, created_at, updated_at
		FROM pages 
		WHERE site_id = (SELECT id FROM sites WHERE code = ?)
		ORDER BY depth, menu_order
	`, siteCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var page models.Page
		if err := rows.Scan(
			&page.ID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
			&page.ParentID, &page.Depth, &page.MenuOrder, &page.Content,
			&page.IsPublished, &page.CreatedAt, &page.UpdatedAt,
		); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pages = append(pages, page)
	}

	json.NewEncoder(w).Encode(pages)
}

func (h *Handler) GetPage(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(chi.URLParam(r, "pageID"))
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	var page models.Page
	err = h.db.QueryRow(`
		SELECT id, site_id, group_id, title, slug, parent_id, depth, 
		menu_order, content, is_published, created_at, updated_at
		FROM pages WHERE id = ?
	`, pageID).Scan(
		&page.ID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
		&page.ParentID, &page.Depth, &page.MenuOrder, &page.Content,
		&page.IsPublished, &page.CreatedAt, &page.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(page)
}

func (h *Handler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(chi.URLParam(r, "pageID"))
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(`
		UPDATE pages 
		SET title = ?, slug = ?, content = ?, updated_at = NOW()
		WHERE id = ?
	`, input.Title, input.Slug, input.Content, pageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"updated": true})
}

func (h *Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(chi.URLParam(r, "pageID"))
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM pages WHERE id = ?", pageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
} 