package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (api *API) listRoutes(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := api.db.Query(r.Context(), `
		SELECT id FROM route
		WHERE published AND ($1 = '' OR title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
		ORDER BY completion_count DESC, updated_at DESC`, query)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	items := make([]routeDTO, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			handleDBError(w, err)
			return
		}
		item, err := api.findRoute(r.Context(), id)
		if err != nil {
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

func (api *API) createRoute(w http.ResponseWriter, r *http.Request) {
	input, ok := readRouteInput(w, r)
	if !ok {
		return
	}
	item, err := api.saveRoute(r.Context(), 0, input)
	if err != nil {
		handleAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (api *API) getRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	item, err := api.findRoute(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (api *API) updateRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	input, ok := readRouteInput(w, r)
	if !ok {
		return
	}
	item, err := api.saveRoute(r.Context(), id, input)
	if err != nil {
		handleAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (api *API) deleteRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	tag, err := api.db.Exec(r.Context(), `DELETE FROM route WHERE id=$1`, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) completeRoute(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	tag, err := api.db.Exec(r.Context(), `UPDATE route SET completion_count=completion_count+1, updated_at=now() WHERE id=$1 AND published`, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "published route not found")
		return
	}
	item, err := api.findRoute(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (api *API) saveRoute(ctx context.Context, id int64, input routeInput) (routeDTO, error) {
	tx, err := api.db.Begin(ctx)
	if err != nil {
		return routeDTO{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if id == 0 {
		err = tx.QueryRow(ctx, `INSERT INTO route (title, description, published, distance_meters, walking_minutes) VALUES ($1, $2, $3, $4, $5) RETURNING id`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), input.Published, input.DistanceMeters, input.WalkingMinutes).Scan(&id)
	} else {
		var tag pgconn.CommandTag
		tag, err = tx.Exec(ctx, `UPDATE route SET title=$1, description=$2, published=$3, distance_meters=$4, walking_minutes=$5, updated_at=now() WHERE id=$6`, strings.TrimSpace(input.Title), strings.TrimSpace(input.Description), input.Published, input.DistanceMeters, input.WalkingMinutes, id)
		if err == nil && tag.RowsAffected() == 0 {
			err = pgx.ErrNoRows
		}
		if err == nil {
			_, err = tx.Exec(ctx, `DELETE FROM route_stop WHERE route_id=$1`, id)
		}
	}
	if err != nil {
		return routeDTO{}, err
	}
	for index, stop := range input.Stops {
		if _, err := tx.Exec(ctx, `INSERT INTO route_stop (route_id, position, location_id, visit_minutes) VALUES ($1, $2, $3, $4)`, id, index+1, stop.LocationID, stop.VisitMinutes); err != nil {
			return routeDTO{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return routeDTO{}, err
	}
	return api.findRoute(ctx, id)
}

func (api *API) findRoute(ctx context.Context, id int64) (routeDTO, error) {
	var item routeDTO
	err := api.db.QueryRow(ctx, `SELECT id, title, description, published, distance_meters, walking_minutes, completion_count, created_at, updated_at FROM route WHERE id=$1`, id).Scan(&item.ID, &item.Title, &item.Description, &item.Published, &item.DistanceMeters, &item.WalkingMinutes, &item.CompletionCount, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return routeDTO{}, err
	}
	rows, err := api.db.Query(ctx, `SELECT s.position, s.location_id, s.visit_minutes, l.id, l.name, l.description, ST_Y(l.geometry), ST_X(l.geometry), l.created_at, l.updated_at FROM route_stop s JOIN location l ON l.id=s.location_id WHERE s.route_id=$1 ORDER BY s.position`, id)
	if err != nil {
		return routeDTO{}, err
	}
	defer rows.Close()
	item.Stops = make([]routeStopDTO, 0)
	for rows.Next() {
		var stop routeStopDTO
		if err := rows.Scan(&stop.Position, &stop.LocationID, &stop.VisitMinutes, &stop.Location.ID, &stop.Location.Name, &stop.Location.Description, &stop.Location.Latitude, &stop.Location.Longitude, &stop.Location.CreatedAt, &stop.Location.UpdatedAt); err != nil {
			return routeDTO{}, err
		}
		item.Stops = append(item.Stops, stop)
	}
	return item, rows.Err()
}
