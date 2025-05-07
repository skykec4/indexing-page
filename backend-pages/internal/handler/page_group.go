package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"pages/internal/models"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetPageGroups(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")

	// 사이트 ID 조회
	var siteID int
	err := h.db.QueryRow("SELECT id FROM sites WHERE code = ?", siteCode).Scan(&siteID)
	if err == sql.ErrNoRows {
		http.Error(w, "사이트를 찾을 수 없습니다", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 페이지 그룹 조회
	rows, err := h.db.Query(
		"SELECT group_id, site_id, name, description, created_at, updated_at FROM page_groups WHERE site_id = ? ORDER BY name",
		siteID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groups []models.PageGroup
	for rows.Next() {
		var group models.PageGroup
		if err := rows.Scan(&group.GroupID, &group.SiteID, &group.Name, &group.Description, &group.CreatedAt, &group.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		groups = append(groups, group)
	}

	json.NewEncoder(w).Encode(groups)
}

func (h *Handler) CreatePageGroup(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")

	// 사이트 ID 조회
	var siteID int
	err := h.db.QueryRow("SELECT id FROM sites WHERE code = ?", siteCode).Scan(&siteID)
	if err == sql.ErrNoRows {
		http.Error(w, "사이트를 찾을 수 없습니다", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var input models.CreatePageGroupInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(
		"INSERT INTO page_groups (site_id, name, description) VALUES (?, ?, ?)",
		siteID, input.Name, input.Description,
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
		"created":  true,
		"group_id": id,
	})
}

func (h *Handler) UpdatePageGroup(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(chi.URLParam(r, "groupId"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec(
		"UPDATE page_groups SET name = ?, description = ? WHERE group_id = ?",
		input.Name, input.Description, groupId,
	)
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
		http.Error(w, "페이지 그룹을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"updated": true})
}

func (h *Handler) DeletePageGroup(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(chi.URLParam(r, "groupId"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("DELETE FROM page_groups WHERE group_id = ?", groupId)
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
		http.Error(w, "페이지 그룹을 찾을 수 없습니다", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"deleted": true})
} 