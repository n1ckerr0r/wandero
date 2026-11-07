# API coverage

`backend/api/openapi.yaml` is the canonical HTTP contract. `make generate` produces the Go router/models and `frontend/src/api-schema.d.ts`. The frontend API client accepts only paths present in the generated `paths` type. `make check` checks the generated schema and runs one frontend client test for every documented operation.

The current OpenAPI contract has 27 operations. The generated Go `ServerInterface` is implemented by `backend/internal/api.API`; the compile-time interface assertion in `contract.go` makes a missing server operation a build error.

| API group | Contract operations | TypeScript client | Current desktop UI |
| --- | --- | --- | --- |
| Health | `GET /health` | Yes | Used by infrastructure health checks |
| Locations | `GET`, `POST /api/locations`; `GET`, `PUT`, `DELETE /api/locations/{id}` | Yes | Calls `GET /api/locations`; place mutations are reserved for data-management tools |
| Routes | `GET`, `POST /api/routes`; `GET`, `PUT`, `DELETE /api/routes/{id}`; `POST /api/routes/{id}/completions` | Yes | Loads published routes, saves a custom ordered walk and records a completion |
| Events and map | `GET`, `POST /api/events`; `GET`, `PUT`, `DELETE /api/events/{id}`; `GET /api/map/events.geojson` | Yes | No event screen in the current MVP |
| Event media | `GET`, `POST /api/events/{id}/media`; `DELETE /api/media/{id}` | Yes | Not shown in the current UI |
| Event sources | `GET /api/events/{id}/sources`; `POST`, `DELETE /api/events/{id}/sources/{source_id}` | Yes | Not shown in the current UI |
| Tags | `GET`, `POST /api/tags`; `PUT`, `DELETE /api/tags/{id}` | Yes | Not shown in the current UI |
| Sources | `GET`, `POST /api/sources`; `GET`, `PUT`, `DELETE /api/sources/{id}` | Yes | Not shown in the current UI |

All current backend operations have a Go handler and a TypeScript client function. Only a subset is called by the desktop screen. The unused CRUD operations remain useful for seed-data maintenance and a future admin/editor, but this repository has no admin UI.

## Routes and routing geometry

Routes are persisted with ordered location stops, optional editor-supplied distance/walking-duration fields and an actual completion counter. The catalogue only exposes published routes; a completion can only be recorded for such a route. The current product deliberately does not infer a walking polyline, distance or duration from straight-line coordinates: a pedestrian-routing provider is still needed before those values or geometry can be computed automatically.

## Write authorization

Mutating operations accept the backend's `API_WRITE_KEY` as a bearer token when that setting is enabled. The frontend does not include the server key in its Vite bundle; `setAPIWriteToken` accepts an in-memory credential for a future authenticated editor. A browser admin must obtain a short-lived credential through an authentication service or a trusted server-side proxy before enabling writes.
