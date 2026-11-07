package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// listEvents возвращает события с опциональными фильтрами по дате, объекту, тегу, тексту, limit и offset.
func (api *API) listEvents(w http.ResponseWriter, r *http.Request) {
	events, err := api.queryEvents(r.Context(), r.URL.Query(), false)
	if err != nil {
		handleAPIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, events)
}

// createEvent создает событие и записывает набор тегов в одной транзакции.
func (api *API) createEvent(w http.ResponseWriter, r *http.Request) {
	input, ok := readEventInput(w, r)
	if !ok {
		return
	}

	item, err := api.insertEvent(r.Context(), input)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// getEvent возвращает одно событие по id из path-параметра.
func (api *API) getEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	item, err := api.findEvent(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// updateEvent заменяет текст события, даты, объект и набор тегов.
func (api *API) updateEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	input, ok := readEventInput(w, r)
	if !ok {
		return
	}

	item, err := api.saveEvent(r.Context(), id, input)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// deleteEvent удаляет событие вместе со связанными тегами, медиа и ссылками на источники.
func (api *API) deleteEvent(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	tag, err := api.db.Exec(r.Context(), `DELETE FROM event WHERE id=$1`, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "event not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) insertEvent(ctx context.Context, input eventInput) (eventDTO, error) {
	dateFrom, dateTo, err := parseEventDates(input)
	if err != nil {
		return eventDTO{}, err
	}

	tx, err := api.db.Begin(ctx)
	if err != nil {
		return eventDTO{}, err
	}
	defer tx.Rollback(ctx)

	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO event (title, description, date_from, date_to, location_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), dateFrom, dateTo, input.LocationID,
	).Scan(&id)
	if err != nil {
		return eventDTO{}, err
	}
	if err := replaceEventTags(ctx, tx, id, input.Tags); err != nil {
		return eventDTO{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return eventDTO{}, err
	}

	return api.findEvent(ctx, id)
}

func (api *API) saveEvent(ctx context.Context, id int64, input eventInput) (eventDTO, error) {
	dateFrom, dateTo, err := parseEventDates(input)
	if err != nil {
		return eventDTO{}, err
	}

	tx, err := api.db.Begin(ctx)
	if err != nil {
		return eventDTO{}, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE event
		SET title=$1, description=$2, date_from=$3, date_to=$4, location_id=$5, updated_at=now()
		WHERE id=$6`,
		strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), dateFrom, dateTo, input.LocationID, id,
	)
	if err != nil {
		return eventDTO{}, err
	}
	if tag.RowsAffected() == 0 {
		return eventDTO{}, pgx.ErrNoRows
	}
	if err := replaceEventTags(ctx, tx, id, input.Tags); err != nil {
		return eventDTO{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return eventDTO{}, err
	}

	return api.findEvent(ctx, id)
}

func (api *API) findEvent(ctx context.Context, id int64) (eventDTO, error) {
	events, err := api.queryEvents(ctx, mapValues{"id": strconv.FormatInt(id, 10)}, true)
	if err != nil {
		return eventDTO{}, err
	}
	if len(events) == 0 {
		return eventDTO{}, pgx.ErrNoRows
	}
	return events[0], nil
}

type mapValues map[string]string

func (m mapValues) Get(key string) string {
	return m[key]
}

func (api *API) queryEvents(ctx context.Context, values interface{ Get(string) string }, exact bool) ([]eventDTO, error) {
	args := make([]any, 0)
	conditions := make([]string, 0)

	if id := values.Get("id"); id != "" {
		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil || parsed <= 0 {
			return nil, badRequestError("id must be a positive integer")
		}
		args = append(args, parsed)
		conditions = append(conditions, fmt.Sprintf("e.id=$%d", len(args)))
	}
	if locationID := values.Get("location_id"); locationID != "" {
		parsed, err := strconv.ParseInt(locationID, 10, 64)
		if err != nil || parsed <= 0 {
			return nil, badRequestError("location_id must be a positive integer")
		}
		args = append(args, parsed)
		conditions = append(conditions, fmt.Sprintf("e.location_id=$%d", len(args)))
	}
	if q := strings.TrimSpace(values.Get("q")); q != "" {
		args = append(args, "%"+q+"%")
		conditions = append(conditions, fmt.Sprintf("(e.title ILIKE $%d OR e.description ILIKE $%d)", len(args), len(args)))
	}
	if tag := strings.TrimSpace(values.Get("tag")); tag != "" {
		args = append(args, tag)
		conditions = append(conditions, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM event_tag et_filter
			JOIN tag t_filter ON t_filter.id=et_filter.tag_id
			WHERE et_filter.event_id=e.id AND t_filter.name=$%d
		)`, len(args)))
	}
	if from := strings.TrimSpace(values.Get("from")); from != "" {
		dateFrom, err := parseFlexibleDate(from, false)
		if err != nil {
			return nil, err
		}
		args = append(args, dateFrom)
		conditions = append(conditions, fmt.Sprintf("COALESCE(e.date_to, e.date_from) >= $%d", len(args)))
	}
	if to := strings.TrimSpace(values.Get("to")); to != "" {
		dateTo, err := parseFlexibleDate(to, true)
		if err != nil {
			return nil, err
		}
		args = append(args, dateTo)
		conditions = append(conditions, fmt.Sprintf("e.date_from <= $%d", len(args)))
	}

	query := `
		SELECT
			e.id,
			e.title,
			e.description,
			e.date_from::text,
			e.date_to::text,
			e.location_id,
			COALESCE(l.name, ''),
			ST_Y(l.geometry),
			ST_X(l.geometry),
			COALESCE(array_agg(t.name ORDER BY t.name) FILTER (WHERE t.id IS NOT NULL), '{}'),
			e.created_at,
			e.updated_at
		FROM event e
		LEFT JOIN location l ON l.id=e.location_id
		LEFT JOIN event_tag et ON et.event_id=e.id
		LEFT JOIN tag t ON t.id=et.tag_id`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += `
		GROUP BY e.id, l.id
		ORDER BY e.date_from, e.id`

	if !exact {
		limit, err := parsePositiveInt(values.Get("limit"), 100)
		if err != nil {
			return nil, err
		}
		offset, err := parseOffset(values.Get("offset"), 0)
		if err != nil {
			return nil, err
		}
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
		args = append(args, offset)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := api.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]eventDTO, 0)
	for rows.Next() {
		var item eventDTO
		var dateTo sql.NullString
		var locationID sql.NullInt64
		var latitude sql.NullFloat64
		var longitude sql.NullFloat64
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.DateFrom,
			&dateTo,
			&locationID,
			&item.LocationName,
			&latitude,
			&longitude,
			&item.Tags,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if dateTo.Valid {
			item.DateTo = &dateTo.String
		}
		if locationID.Valid {
			item.LocationID = &locationID.Int64
		}
		if latitude.Valid {
			item.Latitude = &latitude.Float64
		}
		if longitude.Valid {
			item.Longitude = &longitude.Float64
		}
		events = append(events, item)
	}

	return events, rows.Err()
}
