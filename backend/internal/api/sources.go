package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

// listSources возвращает все исторические источники, отсортированные по названию.
func (api *API) listSources(w http.ResponseWriter, r *http.Request) {
	rows, err := api.db.Query(r.Context(), `SELECT id, title, author, year, url, type::text FROM source ORDER BY title`)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	items, err := scanSources(rows)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// createSource создает источник, который можно привязать к событиям.
func (api *API) createSource(w http.ResponseWriter, r *http.Request) {
	input, ok := readSourceInput(w, r)
	if !ok {
		return
	}

	var id int64
	err := api.db.QueryRow(r.Context(), `
		INSERT INTO source (title, author, year, url, type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		strings.TrimSpace(input.Title), strings.TrimSpace(input.Author), input.Year, strings.TrimSpace(input.URL), input.Type,
	).Scan(&id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	item, err := api.findSource(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

// getSource возвращает один источник по id из path-параметра.
func (api *API) getSource(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	item, err := api.findSource(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// updateSource заменяет метаданные источника.
func (api *API) updateSource(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	input, ok := readSourceInput(w, r)
	if !ok {
		return
	}

	tag, err := api.db.Exec(r.Context(), `
		UPDATE source
		SET title=$1, author=$2, year=$3, url=$4, type=$5
		WHERE id=$6`,
		strings.TrimSpace(input.Title), strings.TrimSpace(input.Author), input.Year, strings.TrimSpace(input.URL), input.Type, id,
	)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "source not found")
		return
	}

	item, err := api.findSource(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// deleteSource удаляет источник и каскадно удаляет связи с событиями.
func (api *API) deleteSource(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	tag, err := api.db.Exec(r.Context(), `DELETE FROM source WHERE id=$1`, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "source not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) findSource(ctx context.Context, id int64) (sourceDTO, error) {
	rows, err := api.db.Query(ctx, `SELECT id, title, author, year, url, type::text FROM source WHERE id=$1`, id)
	if err != nil {
		return sourceDTO{}, err
	}
	defer rows.Close()

	items, err := scanSources(rows)
	if err != nil {
		return sourceDTO{}, err
	}
	if len(items) == 0 {
		return sourceDTO{}, pgx.ErrNoRows
	}
	return items[0], nil
}

// listEventSources возвращает источники, привязанные к одному событию.
func (api *API) listEventSources(w http.ResponseWriter, r *http.Request) {
	eventID, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	rows, err := api.db.Query(r.Context(), `
		SELECT s.id, s.title, s.author, s.year, s.url, s.type::text
		FROM source s
		JOIN event_source es ON es.source_id=s.id
		WHERE es.event_id=$1
		ORDER BY s.title`, eventID)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	items, err := scanSources(rows)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, items)
}

// linkEventSource связывает существующий источник с существующим событием.
func (api *API) linkEventSource(w http.ResponseWriter, r *http.Request) {
	eventID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	sourceID, ok := pathID(w, r, "source_id")
	if !ok {
		return
	}

	_, err := api.db.Exec(r.Context(), `
		INSERT INTO event_source (event_id, source_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, eventID, sourceID)
	if err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// unlinkEventSource удаляет связь источника с событием.
func (api *API) unlinkEventSource(w http.ResponseWriter, r *http.Request) {
	eventID, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	sourceID, ok := pathID(w, r, "source_id")
	if !ok {
		return
	}

	_, err := api.db.Exec(r.Context(), `DELETE FROM event_source WHERE event_id=$1 AND source_id=$2`, eventID, sourceID)
	if err != nil {
		handleDBError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
