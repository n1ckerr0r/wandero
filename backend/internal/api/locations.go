package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

// listLocations возвращает все объекты карты, отсортированные по названию.
func (api *API) listLocations(w http.ResponseWriter, r *http.Request) {
	rows, err := api.db.Query(r.Context(), `
		SELECT id, name, description, ST_Y(geometry), ST_X(geometry), created_at, updated_at
		FROM location
		ORDER BY name`)
	if err != nil {
		handleDBError(w, err)
		return
	}
	defer rows.Close()

	locations := make([]locationDTO, 0)
	for rows.Next() {
		var item locationDTO
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Latitude, &item.Longitude, &item.CreatedAt, &item.UpdatedAt); err != nil {
			handleDBError(w, err)
			return
		}
		locations = append(locations, item)
	}
	if err := rows.Err(); err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, locations)
}

// createLocation создает объект с PostGIS-точкой по широте и долготе.
func (api *API) createLocation(w http.ResponseWriter, r *http.Request) {
	input, ok := readLocationInput(w, r)
	if !ok {
		return
	}

	item, err := api.insertLocation(r.Context(), input)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

// getLocation возвращает один объект по id из path-параметра.
func (api *API) getLocation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	item, err := api.findLocation(r.Context(), id)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// updateLocation заменяет текстовые поля объекта и его координаты.
func (api *API) updateLocation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}
	input, ok := readLocationInput(w, r)
	if !ok {
		return
	}

	item, err := api.saveLocation(r.Context(), id, input)
	if err != nil {
		handleDBError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// deleteLocation удаляет объект и отвязывает связанные события через ON DELETE SET NULL.
func (api *API) deleteLocation(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "id")
	if !ok {
		return
	}

	tag, err := api.db.Exec(r.Context(), `DELETE FROM location WHERE id=$1`, id)
	if err != nil {
		handleDBError(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "location not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (api *API) insertLocation(ctx context.Context, input locationInput) (locationDTO, error) {
	var id int64
	err := api.db.QueryRow(ctx, `
		INSERT INTO location (name, description, geometry)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326))
		RETURNING id`,
		strings.TrimSpace(input.Name), strings.TrimSpace(input.Description), input.Longitude, input.Latitude,
	).Scan(&id)
	if err != nil {
		return locationDTO{}, err
	}

	return api.findLocation(ctx, id)
}

func (api *API) saveLocation(ctx context.Context, id int64, input locationInput) (locationDTO, error) {
	tag, err := api.db.Exec(ctx, `
		UPDATE location
		SET name=$1, description=$2, geometry=ST_SetSRID(ST_MakePoint($3, $4), 4326), updated_at=now()
		WHERE id=$5`,
		strings.TrimSpace(input.Name), strings.TrimSpace(input.Description), input.Longitude, input.Latitude, id,
	)
	if err != nil {
		return locationDTO{}, err
	}
	if tag.RowsAffected() == 0 {
		return locationDTO{}, pgx.ErrNoRows
	}

	return api.findLocation(ctx, id)
}

func (api *API) findLocation(ctx context.Context, id int64) (locationDTO, error) {
	var item locationDTO
	err := api.db.QueryRow(ctx, `
		SELECT id, name, description, ST_Y(geometry), ST_X(geometry), created_at, updated_at
		FROM location
		WHERE id=$1`, id,
	).Scan(&item.ID, &item.Name, &item.Description, &item.Latitude, &item.Longitude, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}
