package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIntegrationLocationCRUD(t *testing.T) {
	handler, _ := newIntegrationAPI(t)

	location := postJSON[locationDTO](t, handler, "/api/locations", `{
		"name":"Адмиралтейство",
		"description":"Главное адмиралтейство",
		"latitude":59.9375,
		"longitude":30.3086
	}`, http.StatusCreated)
	if location.ID == 0 || location.Name != "Адмиралтейство" {
		t.Fatalf("unexpected location: %+v", location)
	}
	got := getJSON[locationDTO](t, handler, "/api/locations/"+itoa(location.ID), http.StatusOK)
	if got.Latitude != location.Latitude || got.Longitude != location.Longitude {
		t.Fatalf("location coordinates were not round-tripped: %+v", got)
	}

	updated := putJSON[locationDTO](t, handler, "/api/locations/"+itoa(location.ID), `{
		"name":"Адмиралтейство обновлено",
		"description":"Обновленное описание",
		"latitude":59.938,
		"longitude":30.309
	}`, http.StatusOK)
	if updated.Name != "Адмиралтейство обновлено" {
		t.Fatalf("location was not updated: %+v", updated)
	}

	getJSON[[]locationDTO](t, handler, "/api/locations", http.StatusOK)
	deleteRequest(t, handler, "/api/locations/"+itoa(location.ID), http.StatusNoContent)
	getRequest(t, handler, "/api/locations/"+itoa(location.ID), http.StatusNotFound)
}

func TestIntegrationRouteOrderedStopsAndCompletion(t *testing.T) {
	handler, _ := newIntegrationAPI(t)
	first := createIntegrationLocation(t, handler)
	second := postJSON[locationDTO](t, handler, "/api/locations", `{"name":"Второе место","description":"","latitude":59.94,"longitude":30.32}`, http.StatusCreated)

	route := postJSON[routeDTO](t, handler, "/api/routes", `{
		"title":"Тестовая прогулка", "description":"Маршрут", "published":true,
		"stops":[
			{"location_id":`+itoa(second.ID)+`,"visit_minutes":15},
			{"location_id":`+itoa(first.ID)+`,"visit_minutes":30}
		]
	}`, http.StatusCreated)
	if len(route.Stops) != 2 || route.Stops[0].LocationID != second.ID || route.Stops[1].Position != 2 {
		t.Fatalf("route stops did not retain order: %+v", route.Stops)
	}

	listed := getJSON[[]routeDTO](t, handler, "/api/routes?q=Тестовая", http.StatusOK)
	if len(listed) != 1 || listed[0].ID != route.ID {
		t.Fatalf("published route is missing from catalogue: %+v", listed)
	}
	completed := postJSON[routeDTO](t, handler, "/api/routes/"+itoa(route.ID)+"/completions", "", http.StatusOK)
	if completed.CompletionCount != 1 {
		t.Fatalf("completion count was not persisted: %+v", completed)
	}

	putJSON[routeDTO](t, handler, "/api/routes/"+itoa(route.ID), `{
		"title":"Тестовая прогулка", "published":true,
		"stops":[{"location_id":`+itoa(first.ID)+`,"visit_minutes":0}]
	}`, http.StatusOK)
	updated := getJSON[routeDTO](t, handler, "/api/routes/"+itoa(route.ID), http.StatusOK)
	if len(updated.Stops) != 1 || updated.Stops[0].LocationID != first.ID {
		t.Fatalf("route stops were not replaced: %+v", updated.Stops)
	}
}

func TestIntegrationTagCRUD(t *testing.T) {
	handler, _ := newIntegrationAPI(t)

	tag := postJSON[tagDTO](t, handler, "/api/tags", `{"name":"integration-tag-created"}`, http.StatusCreated)
	if tag.ID == 0 || tag.Name != "integration-tag-created" {
		t.Fatalf("unexpected tag: %+v", tag)
	}
	duplicate := postJSON[tagDTO](t, handler, "/api/tags", `{"name":"integration-tag-created"}`, http.StatusCreated)
	if duplicate.ID != tag.ID {
		t.Fatalf("tag creation should be idempotent by name: %d != %d", duplicate.ID, tag.ID)
	}

	updated := putJSON[tagDTO](t, handler, "/api/tags/"+itoa(tag.ID), `{"name":"integration-tag-updated"}`, http.StatusOK)
	if updated.Name != "integration-tag-updated" {
		t.Fatalf("tag was not updated: %+v", updated)
	}

	tags := getJSON[[]tagDTO](t, handler, "/api/tags", http.StatusOK)
	if !hasTag(tags, updated.ID, "integration-tag-updated") {
		t.Fatalf("unexpected tags: %+v", tags)
	}

	deleteRequest(t, handler, "/api/tags/"+itoa(tag.ID), http.StatusNoContent)
}

func TestIntegrationSourceCRUDAndEventLinks(t *testing.T) {
	handler, _ := newIntegrationAPI(t)
	event := createIntegrationEvent(t, handler)

	source := postJSON[sourceDTO](t, handler, "/api/sources", `{
		"title":"Архивный источник",
		"author":"ЦГИА СПб",
		"year":1900,
		"url":"https://example.test/source",
		"type":"archive"
	}`, http.StatusCreated)
	gotSource := getJSON[sourceDTO](t, handler, "/api/sources/"+itoa(source.ID), http.StatusOK)
	if gotSource.ID != source.ID || gotSource.Type != "archive" {
		t.Fatalf("unexpected source lookup: %+v", gotSource)
	}

	postRequest(t, handler, "/api/events/"+itoa(event.ID)+"/sources/"+itoa(source.ID), "", http.StatusNoContent)
	linked := getJSON[[]sourceDTO](t, handler, "/api/events/"+itoa(event.ID)+"/sources", http.StatusOK)
	if len(linked) != 1 || linked[0].ID != source.ID {
		t.Fatalf("unexpected linked sources: %+v", linked)
	}

	updated := putJSON[sourceDTO](t, handler, "/api/sources/"+itoa(source.ID), `{
		"title":"Обновленный источник",
		"author":"ЦГИА СПб",
		"year":1901,
		"url":"https://example.test/source-updated",
		"type":"archive"
	}`, http.StatusOK)
	if updated.Title != "Обновленный источник" {
		t.Fatalf("source was not updated: %+v", updated)
	}

	deleteRequest(t, handler, "/api/events/"+itoa(event.ID)+"/sources/"+itoa(source.ID), http.StatusNoContent)
	deleteRequest(t, handler, "/api/sources/"+itoa(source.ID), http.StatusNoContent)
}

func TestIntegrationEventCRUDAndGeoJSON(t *testing.T) {
	handler, _ := newIntegrationAPI(t)
	location := createIntegrationLocation(t, handler)

	event := postJSON[eventDTO](t, handler, "/api/events", `{
		"title":"Основание Санкт-Петербурга",
		"description":"Начало строительства города",
		"date_from":"1703-05-27",
		"location_id":`+itoa(location.ID)+`,
		"tags":["основание","XVIII век"]
	}`, http.StatusCreated)
	if event.ID == 0 || event.LocationID == nil || *event.LocationID != location.ID {
		t.Fatalf("unexpected event: %+v", event)
	}
	gotEvent := getJSON[eventDTO](t, handler, "/api/events/"+itoa(event.ID), http.StatusOK)
	if gotEvent.ID != event.ID || gotEvent.LocationName != location.Name || len(gotEvent.Tags) != 2 {
		t.Fatalf("unexpected event lookup: %+v", gotEvent)
	}

	events := getJSON[[]eventDTO](t, handler, "/api/events?from=1700&to=1710&tag=основание", http.StatusOK)
	if !hasEvent(events, event.ID) {
		t.Fatalf("unexpected filtered events: %+v", events)
	}

	geojson := getJSON[map[string]any](t, handler, "/api/map/events.geojson?from=1700&to=1710", http.StatusOK)
	if geojson["type"] != "FeatureCollection" {
		t.Fatalf("unexpected geojson: %+v", geojson)
	}
	features, ok := geojson["features"].([]any)
	if !ok {
		t.Fatalf("expected GeoJSON features array, got: %#v", geojson["features"])
	}
	foundEvent := false
	for _, rawFeature := range features {
		feature, ok := rawFeature.(map[string]any)
		if !ok {
			continue
		}
		properties, ok := feature["properties"].(map[string]any)
		if ok && properties["id"] == float64(event.ID) {
			foundEvent = true
			break
		}
	}
	if !foundEvent {
		t.Fatalf("GeoJSON does not contain created event %d", event.ID)
	}

	updated := putJSON[eventDTO](t, handler, "/api/events/"+itoa(event.ID), `{
		"title":"Основание города",
		"description":"Обновленное описание",
		"date_from":"1703-05-27",
		"location_id":`+itoa(location.ID)+`,
		"tags":["основание"]
	}`, http.StatusOK)
	if updated.Title != "Основание города" || len(updated.Tags) != 1 {
		t.Fatalf("event was not updated: %+v", updated)
	}

	deleteRequest(t, handler, "/api/events/"+itoa(event.ID), http.StatusNoContent)
}

func TestIntegrationHealth(t *testing.T) {
	handler, _ := newIntegrationAPI(t)
	status := getJSON[map[string]string](t, handler, "/health", http.StatusOK)
	if status["status"] != "ok" {
		t.Fatalf("unexpected health response: %#v", status)
	}
}

func TestIntegrationMediaCRUD(t *testing.T) {
	handler, _ := newIntegrationAPI(t)
	event := createIntegrationEvent(t, handler)

	media := postJSON[mediaDTO](t, handler, "/api/events/"+itoa(event.ID)+"/media", `{
		"url":"https://example.test/photo.jpg",
		"caption":"Фото",
		"type":"image"
	}`, http.StatusCreated)
	if media.ID == 0 || media.EventID != event.ID {
		t.Fatalf("unexpected media: %+v", media)
	}

	items := getJSON[[]mediaDTO](t, handler, "/api/events/"+itoa(event.ID)+"/media", http.StatusOK)
	if len(items) != 1 || items[0].ID != media.ID {
		t.Fatalf("unexpected media list: %+v", items)
	}

	deleteRequest(t, handler, "/api/media/"+itoa(media.ID), http.StatusNoContent)
}

func newIntegrationAPI(t *testing.T) (http.Handler, *pgxpool.Pool) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	resetIntegrationDatabase(t, pool)
	return New(pool, nil, ""), pool
}

func resetIntegrationDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS media, event_source, source, event_tag, tag, event, location CASCADE;
		DROP TYPE IF EXISTS source_type CASCADE;
	`)
	if err != nil {
		t.Fatalf("reset test database: %v", err)
	}

	migrationDir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		sqlBytes, err := os.ReadFile(filepath.Join(migrationDir, entry.Name()))
		if err != nil {
			t.Fatalf("read migration %s: %v", entry.Name(), err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			t.Fatalf("apply migration %s: %v", entry.Name(), err)
		}
	}
}

func createIntegrationLocation(t *testing.T, handler http.Handler) locationDTO {
	t.Helper()
	return postJSON[locationDTO](t, handler, "/api/locations", `{
		"name":"Петропавловская крепость",
		"description":"Исторический центр",
		"latitude":59.95,
		"longitude":30.3167
	}`, http.StatusCreated)
}

func createIntegrationEvent(t *testing.T, handler http.Handler) eventDTO {
	t.Helper()
	location := createIntegrationLocation(t, handler)
	return postJSON[eventDTO](t, handler, "/api/events", `{
		"title":"Тестовое событие",
		"description":"Описание",
		"date_from":"1703-05-27",
		"location_id":`+itoa(location.ID)+`,
		"tags":["test"]
	}`, http.StatusCreated)
}

func postJSON[T any](t *testing.T, handler http.Handler, path string, body string, expectedStatus int) T {
	t.Helper()
	return jsonRequest[T](t, handler, http.MethodPost, path, body, expectedStatus)
}

func putJSON[T any](t *testing.T, handler http.Handler, path string, body string, expectedStatus int) T {
	t.Helper()
	return jsonRequest[T](t, handler, http.MethodPut, path, body, expectedStatus)
}

func getJSON[T any](t *testing.T, handler http.Handler, path string, expectedStatus int) T {
	t.Helper()
	return jsonRequest[T](t, handler, http.MethodGet, path, "", expectedStatus)
}

func getRequest(t *testing.T, handler http.Handler, path string, expectedStatus int) {
	t.Helper()
	request(t, handler, http.MethodGet, path, "", expectedStatus)
}

func putRequest(t *testing.T, handler http.Handler, path string, body string, expectedStatus int) {
	t.Helper()
	request(t, handler, http.MethodPut, path, body, expectedStatus)
}

func postRequest(t *testing.T, handler http.Handler, path string, body string, expectedStatus int) {
	t.Helper()
	request(t, handler, http.MethodPost, path, body, expectedStatus)
}

func deleteRequest(t *testing.T, handler http.Handler, path string, expectedStatus int) {
	t.Helper()
	request(t, handler, http.MethodDelete, path, "", expectedStatus)
}

func jsonRequest[T any](t *testing.T, handler http.Handler, method string, path string, body string, expectedStatus int) T {
	t.Helper()
	rec := request(t, handler, method, path, body, expectedStatus)

	var output T
	if err := json.Unmarshal(rec.Body.Bytes(), &output); err != nil {
		t.Fatalf("decode %s %s response: %v\nbody: %s", method, path, err, rec.Body.String())
	}
	return output
}

func request(t *testing.T, handler http.Handler, method string, path string, body string, expectedStatus int) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != expectedStatus {
		t.Fatalf("%s %s returned status %d, want %d; body=%s", method, path, rec.Code, expectedStatus, rec.Body.String())
	}
	return rec
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}

func hasTag(tags []tagDTO, id int64, name string) bool {
	for _, tag := range tags {
		if tag.ID == id && tag.Name == name {
			return true
		}
	}
	return false
}

func hasEvent(events []eventDTO, id int64) bool {
	for _, event := range events {
		if event.ID == id {
			return true
		}
	}
	return false
}
