package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"pages/internal/models"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// GetSites godoc
// @Summary 모든 사이트 목록 조회
// @Description 등록된 모든 사이트의 목록을 조회합니다.
// @Tags sites
// @Accept json
// @Produce json
// @Success 200 {array} models.Site
// @Router /api/sites [get]
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
		if err := rows.Scan(&site.SiteID, &site.Code, &site.Name, &site.Domain, &site.CreatedAt, &site.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sites = append(sites, site)
	}

	json.NewEncoder(w).Encode(sites)
}

// CreateSite godoc
// @Summary 사이트 생성
// @Description 새로운 사이트를 생성합니다.
// @Tags sites
// @Accept json
// @Produce json
// @Param site body models.CreateSiteInput true "Site Info"
// @Success 201 {object} map[string]int
// @Failure 400 {object} map[string]string
// @Router /api/sites [post]
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

// GetSiteMenu godoc
// @Summary 전체 메뉴 조회
// @Description 사이트의 전체 메뉴를 조회합니다.
// @Tags menu
// @Accept json
// @Produce json
// @Param site_code path string true "Site Code"
// @Success 200 {array} models.Page
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/menu [get]
func (h *Handler) GetSiteMenu(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")
	fmt.Println("siteCode : ", siteCode)

	// 사이트 정보 조회
	var site models.Site
	err := h.db.QueryRow(
		"SELECT site_id, code, name, domain, created_at, updated_at FROM sites WHERE code = ?",
		siteCode,
	).Scan(&site.SiteID, &site.Code, &site.Name, &site.Domain, &site.CreatedAt, &site.UpdatedAt)
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
		site.SiteID,
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
			`SELECT page_id, site_id, group_id, title, slug, parent_id, depth, menu_order, 
			content, is_published, created_at, updated_at 
			FROM pages 
			WHERE site_id = ? AND group_id = ? 
			ORDER BY depth, menu_order`,
			site.SiteID, group.GroupID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer pageRows.Close()

		pages := []*models.Page{} // ← 빈 슬라이스로 초기화

		for pageRows.Next() {
			var page models.Page
			if err := pageRows.Scan(
				&page.PageID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
				&page.ParentID, &page.Depth, &page.MenuOrder, &page.Content,
				&page.IsPublished, &page.CreatedAt, &page.UpdatedAt,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Println("page : ", page)

			pages = append(pages, &page)
		}

		fmt.Println("pages : ", len(pages))

		group.Menu = BuildMenuTree(pages)
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

// CreatePage godoc
// @Summary 페이지 생성
// @Description 페이지 생성
// @Tags pages
// @Accept json
// @Produce json
// @Param site_code path string true "Site Code"
// @Param group_id path int true "Group ID"
// @Param page body models.CreatePageInput true "Page Information"
// @Success 201 {object} models.Page
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/groups/{group_id}/pages [post]
func (h *Handler) CreatePage(w http.ResponseWriter, r *http.Request) {
	var input models.CreatePageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	siteCode := chi.URLParam(r, "siteCode")
	groupId, err := strconv.Atoi(chi.URLParam(r, "groupId"))
	if err != nil {
		http.Error(w, "Invalid group_id", http.StatusBadRequest)
		return
	}

	// 사이트 ID 조회
	var siteID int
	err2 := h.db.QueryRow("SELECT site_id FROM sites WHERE code = ?", siteCode).Scan(&siteID)
	if err2 == sql.ErrNoRows {
		http.Error(w, "사이트를 찾을 수 없습니다", http.StatusNotFound)
		return
	} else if err2 != nil {
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}

	// 부모 페이지가 있는 경우 depth 계산
	depth := 0
	// if input.ParentID != nil {
	// 	err := h.db.QueryRow("SELECT depth FROM pages WHERE id = ?", *input.ParentID).Scan(&depth)
	// 	if err != nil {
	// 		http.Error(w, "부모 페이지를 찾을 수 없습니다", http.StatusBadRequest)
	// 		return
	// 	}
	// 	depth++
	// }
	var parentID *int
	if input.ParentID != nil && *input.ParentID != 0 {
		parentID = input.ParentID
	}

	result, err := h.db.Exec(
		`INSERT INTO pages (site_id, group_id, title, slug, parent_id, depth, menu_order, content, is_published) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		siteID, groupId, input.Title, input.Slug, parentID, depth, 0, input.Content, true,
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

	fmt.Println("마지막 생성!!!", id)

	// 생성된 페이지 조회
	var page models.Page

	err = h.db.QueryRow(
		"SELECT page_id, site_id, group_id, title, slug, parent_id, depth, menu_order, content, is_published, created_at, updated_at FROM pages WHERE id = ?", id,
	).Scan(
		&page.PageID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
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

func BuildMenuTree(pages []*models.Page) []*models.Page {
	pageMap := make(map[int]*models.Page)

	// 모든 페이지를 맵에 저장
	for _, page := range pages {
		page.Menu = []*models.Page{} // 초기화
		pageMap[page.PageID] = page
	}

	// 트리 구성
	var roots []*models.Page
	for _, page := range pages {
		if page.ParentID != nil {
			if parent, exists := pageMap[*page.ParentID]; exists {
				parent.Menu = append(parent.Menu, page)
			}
		} else {
			roots = append(roots, page)
		}
	}

	// 디버깅 출력 (선택적)
	for _, page := range pages {
		fmt.Printf("Page ID: %d, Parent ID: %v, Children: %d\n", page.PageID, page.ParentID, len(page.Menu))
	}
	printChildren(roots, 0)

	return roots
}

func printChildren(pages []*models.Page, level int) {
	for _, page := range pages {
		indent := strings.Repeat("  ", level)
		fmt.Printf("%sPage ID: %d, Children: %d\n", indent, page.PageID, len(page.Menu))
		if len(page.Menu) > 0 {
			printChildren(page.Menu, level+1)
		}
	}
}

// ListPages godoc
// @Summary List all pages
// @Description Retrieve a list of all pages for a site and group
// @Tags pages
// @Accept json
// @Produce json
// @Param site_code path string true "Site Code"
// @Param group_id path int true "Group ID"
// @Success 200 {array} models.Page
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/groups/{group_id}/pages [get]
func (h *Handler) ListPages(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")
	groupId, err := strconv.Atoi(chi.URLParam(r, "groupId"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	rows, err := h.db.Query(`
		SELECT page_id, site_id, group_id, title, slug, parent_id, depth, 
		menu_order, content, is_published, created_at, updated_at
		FROM pages 
		WHERE site_id = (SELECT site_id FROM sites WHERE code = ?)
		AND group_id = ?
		ORDER BY depth, menu_order
	`, siteCode, groupId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pages []models.Page
	for rows.Next() {
		var page models.Page
		if err := rows.Scan(
			&page.PageID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
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

// GetPage godoc
// @Summary Get page by ID
// @Description Retrieve a specific page by its ID
// @Tags pages
// @Accept json
// @Produce json
// @Param site_code path string true "Site Code"
// @Param page_id path int true "Page ID"
// @Success 200 {object} models.Page
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/pages/{page_id} [get]
func (h *Handler) GetPage(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(chi.URLParam(r, "pageID"))
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	var page models.Page
	err = h.db.QueryRow(`
		SELECT page_id, site_id, group_id, title, slug, parent_id, depth, 
		menu_order, content, is_published, created_at, updated_at
		FROM pages WHERE page_id = ?
	`, pageID).Scan(
		&page.PageID, &page.SiteID, &page.GroupID, &page.Title, &page.Slug,
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

// UpdatePage godoc
// @Summary Update page
// @Description Update an existing page with new information
// @Tags pages
// @Accept json
// @Produce json
// @Param site_code path string true "Site Code"
// @Param page_id path int true "Page ID"
// @Param page body models.UpdatePageInput true "Page Information"
// @Success 200 {object} models.Page
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/pages/{page_id} [put]
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
		WHERE page_id = ?
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

// DeletePage godoc
// @Summary Delete page
// @Description Delete a specific page
// @Tags pages
// @Accept json
// @Produce json
// @Param site_code path string true "Site Code"
// @Param page_id path int true "Page ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/pages/{page_id} [delete]
func (h *Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	pageID, err := strconv.Atoi(chi.URLParam(r, "pageID"))
	if err != nil {
		http.Error(w, "Invalid page ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM pages WHERE page_id = ?", pageID)
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
