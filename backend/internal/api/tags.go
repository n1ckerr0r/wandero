package api

import (
	"net/http"
	"strings"
)

// listTags возвращает все теги событий, отсортированные по названию.
func (api *API) listTags(w http.ResponseWriter, r *http.Request) {
	rows, err := api.db.Query(r.Context(), `SELECT id, name FROM tag ORDER BY name`)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	tags := make([]tagDTO, 0)
	for rows.Next() {
		var item tagDTO
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			handleDBError(w, err)
			return
		}
		tags = append(tags, item)
	}
	if err := rows.Err(); err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, tags)
}

// createTag создает тег или возвращает существующий тег с таким же названием.
func (api *API) createTag(w http.ResponseWriter, r *http.Request) {
	var input tagInput
	if !readJSONBody(w, r, &input) {
		return
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var item tagDTO
	err := api.db.QueryRow(r.Context(), `
		INSERT INTO tag (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
		RETURNING id, name`, name,
	).Scan(&item.ID, &item.Name)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// updateTag переименовывает существующий тег.
func (api *API) updateTag(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var input tagInput
	if !readJSONBody(w, r, &input) {
		return
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	var item tagDTO
	err := api.db.QueryRow(r.Context(), `UPDATE tag SET name=$1 WHERE id=$2 RETURNING id, name`, name, id).Scan(&item.ID, &item.Name)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// deleteTag удаляет тег и каскадно удаляет его связи с событиями.
func (api *API) deleteTag(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	tag, err := api.db.Exec(r.Context(), `DELETE FROM tag WHERE id=$1`, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "tag not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
