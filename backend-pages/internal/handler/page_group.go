package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"pages/internal/models"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetPageGroups godoc
// @Summary 페이지 그룹 목록 조회
// @Description 사이트에 등록된 모든 페이지 그룹의 목록을 조회합니다.
// @Tags page_groups
// @Accept json
// @Produce json
// @Param site_code path string true "사이트 코드"
// @Success 200 {array} models.PageGroup
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/groups [get]
func (h *Handler) GetPageGroups(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")

	// 사이트 ID 조회
	var siteID int
	err := h.db.QueryRow("SELECT site_id FROM sites WHERE code = ?", siteCode).Scan(&siteID)
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

// CreatePageGroup godoc
// @Summary 페이지 그룹 생성
// @Description 사이트에 새로운 페이지 그룹을 생성합니다.
// @Tags page_groups
// @Accept json
// @Produce json
// @Param siteCode path string true "사이트 코드"
// @Param input body models.CreatePageGroupInput true "페이지 그룹 생성 입력"
// @Success 201 {object} models.PageGroup
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{siteCode}/groups [post]
func (h *Handler) CreatePageGroup(w http.ResponseWriter, r *http.Request) {
	siteCode := chi.URLParam(r, "siteCode")

	// 사이트 ID 조회
	var siteID int
	err := h.db.QueryRow("SELECT site_id FROM sites WHERE code = ?", siteCode).Scan(&siteID)
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

// UpdatePageGroup godoc
// @Summary 페이지 그룹 업데이트
// @Description 사이트의 페이지 그룹 정보를 업데이트합니다.
// @Tags page_groups
// @Accept json
// @Produce json
// @Param site_code path string true "사이트 코드"
// @Param group_id path int true "Group ID"
// @Param input body models.UpdatePageGroupInput true "Page Group Update Input"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/groups/{group_id} [put]
func (h *Handler) UpdatePageGroup(w http.ResponseWriter, r *http.Request) {
	groupId, err := strconv.Atoi(chi.URLParam(r, "groupId"))
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	var input models.UpdatePageGroupInput
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

// DeletePageGroup godoc
// @Summary 페이지 그룹 삭제
// @Description 사이트의 페이지 그룹을 삭제합니다.
// @Tags page_groups
// @Accept json
// @Produce json
// @Param site_code path string true "사이트 코드"
// @Param group_id path int true "Group ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/sites/{site_code}/groups/{group_id} [delete]
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
