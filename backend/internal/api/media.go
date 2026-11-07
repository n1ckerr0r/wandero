package api

import (
	"net/http"
	"net/url"
	"strings"
)

// listEventMedia возвращает медиа, привязанные к одному событию.
func (api *API) listEventMedia(w http.ResponseWriter, r *http.Request) {
	eventID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	rows, err := api.db.Query(r.Context(), `
		SELECT id, event_id, url, caption, type
		FROM media
		WHERE event_id=$1
		ORDER BY id`, eventID)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	items := make([]mediaDTO, 0)
	for rows.Next() {
		var item mediaDTO
		if err := rows.Scan(&item.ID, &item.EventID, &item.URL, &item.Caption, &item.Type); err != nil {
			handleDBError(w, err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// createEventMedia привязывает медиа к событию.
func (api *API) createEventMedia(w http.ResponseWriter, r *http.Request) {
	eventID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	var input mediaInput
	if !readJSONBody(w, r, &input) {
		return
	}
	if strings.TrimSpace(input.URL) == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}
	if strings.TrimSpace(input.Type) == "" {
		input.Type = "image"
	}
	if input.Type != "image" && input.Type != "link" {
		writeError(w, http.StatusBadRequest, "type must be image or link")
		return
	}
	parsedURL, err := url.ParseRequestURI(input.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		writeError(w, http.StatusBadRequest, "url must be an absolute HTTP(S) URL")
		return
	}

	var item mediaDTO
	err = api.db.QueryRow(r.Context(), `
		INSERT INTO media (event_id, url, caption, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, event_id, url, caption, type`,
		eventID, strings.TrimSpace(input.URL), strings.TrimSpace(input.Caption), strings.TrimSpace(input.Type),
	).Scan(&item.ID, &item.EventID, &item.URL, &item.Caption, &item.Type)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// deleteMedia удаляет медиа по id из path-параметра.
func (api *API) deleteMedia(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	tag, err := api.db.Exec(r.Context(), `DELETE FROM media WHERE id=$1`, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "media not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
