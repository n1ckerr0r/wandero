package api

import (
	"net/http"

	"github.com/n1ckerr0r/chronograph-spb/backend/internal/openapi"
)

var _ openapi.ServerInterface = (*API)(nil)

// GetHealth реализует OpenAPI-операцию GET /health.
func (api *API) GetHealth(w http.ResponseWriter, r *http.Request) {
	api.health(w, r)
}

// GetApiLocations реализует OpenAPI-операцию GET /api/locations.
func (api *API) GetApiLocations(w http.ResponseWriter, r *http.Request) {
	api.listLocations(w, r)
}

// PostApiLocations реализует OpenAPI-операцию POST /api/locations.
func (api *API) PostApiLocations(w http.ResponseWriter, r *http.Request) {
	api.createLocation(w, r)
}

// GetApiLocationsId реализует OpenAPI-операцию GET /api/locations/{id}.
func (api *API) GetApiLocationsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.getLocation(w, r)
}

// PutApiLocationsId реализует OpenAPI-операцию PUT /api/locations/{id}.
func (api *API) PutApiLocationsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.updateLocation(w, r)
}

// DeleteApiLocationsId реализует OpenAPI-операцию DELETE /api/locations/{id}.
func (api *API) DeleteApiLocationsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.deleteLocation(w, r)
}

func (api *API) GetApiRoutes(w http.ResponseWriter, r *http.Request, params openapi.GetApiRoutesParams) {
	api.listRoutes(w, r)
}

func (api *API) PostApiRoutes(w http.ResponseWriter, r *http.Request) {
	api.createRoute(w, r)
}

func (api *API) GetApiRoutesId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.getRoute(w, r)
}

func (api *API) PutApiRoutesId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.updateRoute(w, r)
}

func (api *API) DeleteApiRoutesId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.deleteRoute(w, r)
}

func (api *API) PostApiRoutesIdCompletions(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.completeRoute(w, r)
}

// GetApiEvents реализует OpenAPI-операцию GET /api/events.
func (api *API) GetApiEvents(w http.ResponseWriter, r *http.Request, params openapi.GetApiEventsParams) {
	api.listEvents(w, r)
}

// PostApiEvents реализует OpenAPI-операцию POST /api/events.
func (api *API) PostApiEvents(w http.ResponseWriter, r *http.Request) {
	api.createEvent(w, r)
}

// GetApiEventsId реализует OpenAPI-операцию GET /api/events/{id}.
func (api *API) GetApiEventsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.getEvent(w, r)
}

// PutApiEventsId реализует OpenAPI-операцию PUT /api/events/{id}.
func (api *API) PutApiEventsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.updateEvent(w, r)
}

// DeleteApiEventsId реализует OpenAPI-операцию DELETE /api/events/{id}.
func (api *API) DeleteApiEventsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.deleteEvent(w, r)
}

// GetApiEventsIdMedia реализует OpenAPI-операцию GET /api/events/{id}/media.
func (api *API) GetApiEventsIdMedia(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.listEventMedia(w, r)
}

// PostApiEventsIdMedia реализует OpenAPI-операцию POST /api/events/{id}/media.
func (api *API) PostApiEventsIdMedia(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.createEventMedia(w, r)
}

// GetApiEventsIdSources реализует OpenAPI-операцию GET /api/events/{id}/sources.
func (api *API) GetApiEventsIdSources(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.listEventSources(w, r)
}

// PostApiEventsIdSourcesSourceId реализует OpenAPI-операцию POST /api/events/{id}/sources/{source_id}.
func (api *API) PostApiEventsIdSourcesSourceId(w http.ResponseWriter, r *http.Request, id openapi.ID, sourceID int64) {
	api.linkEventSource(w, r)
}

// DeleteApiEventsIdSourcesSourceId реализует OpenAPI-операцию DELETE /api/events/{id}/sources/{source_id}.
func (api *API) DeleteApiEventsIdSourcesSourceId(w http.ResponseWriter, r *http.Request, id openapi.ID, sourceID int64) {
	api.unlinkEventSource(w, r)
}

// GetApiTags реализует OpenAPI-операцию GET /api/tags.
func (api *API) GetApiTags(w http.ResponseWriter, r *http.Request) {
	api.listTags(w, r)
}

// PostApiTags реализует OpenAPI-операцию POST /api/tags.
func (api *API) PostApiTags(w http.ResponseWriter, r *http.Request) {
	api.createTag(w, r)
}

// PutApiTagsId реализует OpenAPI-операцию PUT /api/tags/{id}.
func (api *API) PutApiTagsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.updateTag(w, r)
}

// DeleteApiTagsId реализует OpenAPI-операцию DELETE /api/tags/{id}.
func (api *API) DeleteApiTagsId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.deleteTag(w, r)
}

// GetApiSources реализует OpenAPI-операцию GET /api/sources.
func (api *API) GetApiSources(w http.ResponseWriter, r *http.Request) {
	api.listSources(w, r)
}

// PostApiSources реализует OpenAPI-операцию POST /api/sources.
func (api *API) PostApiSources(w http.ResponseWriter, r *http.Request) {
	api.createSource(w, r)
}

// GetApiSourcesId реализует OpenAPI-операцию GET /api/sources/{id}.
func (api *API) GetApiSourcesId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.getSource(w, r)
}

// PutApiSourcesId реализует OpenAPI-операцию PUT /api/sources/{id}.
func (api *API) PutApiSourcesId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.updateSource(w, r)
}

// DeleteApiSourcesId реализует OpenAPI-операцию DELETE /api/sources/{id}.
func (api *API) DeleteApiSourcesId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.deleteSource(w, r)
}

// DeleteApiMediaId реализует OpenAPI-операцию DELETE /api/media/{id}.
func (api *API) DeleteApiMediaId(w http.ResponseWriter, r *http.Request, id openapi.ID) {
	api.deleteMedia(w, r)
}

// GetApiMapEventsGeojson реализует OpenAPI-операцию GET /api/map/events.geojson.
func (api *API) GetApiMapEventsGeojson(w http.ResponseWriter, r *http.Request, params openapi.GetApiMapEventsGeojsonParams) {
	api.eventsGeoJSON(w, r)
}
