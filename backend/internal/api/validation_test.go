package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseEventDates(t *testing.T) {
	dateTo := "1703-06-01"
	dateFrom, parsedTo, err := parseEventDates(eventInput{
		Title:    "Founded",
		DateFrom: "1703-05-27",
		DateTo:   &dateTo,
	})
	if err != nil {
		t.Fatalf("parseEventDates returned error: %v", err)
	}
	if dateFrom.Format(dateLayout) != "1703-05-27" {
		t.Fatalf("unexpected date_from: %s", dateFrom.Format(dateLayout))
	}
	if parsedTo == nil || parsedTo.Format(dateLayout) != "1703-06-01" {
		t.Fatalf("unexpected date_to: %v", parsedTo)
	}
}

func TestParseEventDatesRejectsInvalidRange(t *testing.T) {
	dateTo := "1703-05-26"
	_, _, err := parseEventDates(eventInput{DateFrom: "1703-05-27", DateTo: &dateTo})
	if err == nil {
		t.Fatal("expected invalid date range error")
	}
}

func TestParseFlexibleDate(t *testing.T) {
	start, err := parseFlexibleDate("1703", false)
	if err != nil {
		t.Fatalf("parseFlexibleDate returned error: %v", err)
	}
	if !start.Equal(time.Date(1703, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected start date: %v", start)
	}

	end, err := parseFlexibleDate("1703", true)
	if err != nil {
		t.Fatalf("parseFlexibleDate returned error: %v", err)
	}
	if !end.Equal(time.Date(1703, time.December, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected end date: %v", end)
	}
}

func TestParseFlexibleDateRejectsInvalidValue(t *testing.T) {
	_, err := parseFlexibleDate("17xx", false)
	if err == nil {
		t.Fatal("expected invalid date filter error")
	}
	if _, err := parseFlexibleDate("1703-02-30", false); err == nil {
		t.Fatal("expected invalid calendar date to be rejected")
	}
}

func TestParsePositiveInt(t *testing.T) {
	if got, err := parsePositiveInt("", 25); got != 25 || err != nil {
		t.Fatalf("empty value returned %d", got)
	}
	if _, err := parsePositiveInt("-1", 25); err == nil {
		t.Fatal("negative value should fail")
	}
	if _, err := parsePositiveInt("501", 25); err == nil {
		t.Fatal("large value should fail")
	}
	if got, err := parsePositiveInt("10", 25); got != 10 || err != nil {
		t.Fatalf("valid value returned %d", got)
	}
}

func TestParseOffset(t *testing.T) {
	if got, err := parseOffset("", 10); got != 10 || err != nil {
		t.Fatalf("empty value returned %d, %v", got, err)
	}
	if got, err := parseOffset("501", 0); got != 501 || err != nil {
		t.Fatalf("valid deep offset returned %d, %v", got, err)
	}
	for _, value := range []string{"-1", "100001", "abc"} {
		if _, err := parseOffset(value, 0); err == nil {
			t.Errorf("parseOffset(%q) should fail", value)
		}
	}
}

func TestIsSourceType(t *testing.T) {
	for _, sourceType := range []string{"book", "article", "archive", "photo", "map", "web"} {
		if !isSourceType(sourceType) {
			t.Fatalf("%s should be accepted", sourceType)
		}
	}
	if isSourceType("video") {
		t.Fatal("unexpected source type accepted")
	}
}

func TestReadLocationInput(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/locations", strings.NewReader(`{
		"name":"Test",
		"description":"Description",
		"latitude":59.95,
		"longitude":30.31
	}`))
	rec := httptest.NewRecorder()

	input, ok := readLocationInput(rec, req)
	if !ok {
		t.Fatalf("readLocationInput failed: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if input.Name != "Test" || input.Latitude != 59.95 || input.Longitude != 30.31 {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestReadLocationInputRejectsInvalidLatitude(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/locations", strings.NewReader(`{
		"name":"Test",
		"latitude":100,
		"longitude":30.31
	}`))
	rec := httptest.NewRecorder()

	if _, ok := readLocationInput(rec, req); ok {
		t.Fatal("expected invalid latitude to be rejected")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestReadEventInput(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/events", strings.NewReader(`{
		"title":"Event",
		"description":"Description",
		"date_from":"1703-05-27",
		"tags":["tag"]
	}`))
	rec := httptest.NewRecorder()

	input, ok := readEventInput(rec, req)
	if !ok {
		t.Fatalf("readEventInput failed: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if input.Title != "Event" || input.DateFrom != "1703-05-27" || len(input.Tags) != 1 {
		t.Fatalf("unexpected input: %+v", input)
	}
}

func TestReadSourceInputDefaultsType(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/sources", strings.NewReader(`{
		"title":"Source",
		"author":"Author"
	}`))
	rec := httptest.NewRecorder()

	input, ok := readSourceInput(rec, req)
	if !ok {
		t.Fatalf("readSourceInput failed: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if input.Type != "web" {
		t.Fatalf("unexpected default source type: %s", input.Type)
	}
}

func TestReadJSONBodyRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/tags", strings.NewReader(`{"name":"tag","extra":true}`))
	rec := httptest.NewRecorder()

	var input tagInput
	if readJSONBody(rec, req, &input) {
		t.Fatal("expected unknown JSON field to be rejected")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestReadJSONBodyRequiresOneValidObjectAndEnforcesLimit(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{"name":`},
		{name: "second JSON value", body: `{"name":"tag"} {"name":"other"}`},
		{name: "trailing garbage", body: `{"name":"tag"} trailing`},
		{name: "oversized JSON", body: `{"name":"` + strings.Repeat("x", (1<<20)+1) + `"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/tags", strings.NewReader(test.body))
			rec := httptest.NewRecorder()
			var input tagInput
			if readJSONBody(rec, req, &input) {
				t.Fatal("invalid body was accepted")
			}
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status %d", rec.Code)
			}
		})
	}
}

func TestPathID(t *testing.T) {
	for _, test := range []struct {
		value string
		want  int64
		ok    bool
	}{
		{value: "42", want: 42, ok: true},
		{value: "0"},
		{value: "-1"},
		{value: "abc"},
	} {
		t.Run(test.value, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/locations/"+test.value, nil)
			req.SetPathValue("id", test.value)
			rec := httptest.NewRecorder()
			got, ok := pathID(rec, req, "id")
			if ok != test.ok || (ok && got != test.want) {
				t.Fatalf("pathID() = %d, %v; want %d, %v", got, ok, test.want, test.ok)
			}
			if !ok && rec.Code != http.StatusBadRequest {
				t.Fatalf("invalid id returned status %d", rec.Code)
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight should not call next handler")
	}), []string{"https://example.test"})
	req := httptest.NewRequest(http.MethodOptions, "/api/events", nil)
	req.Header.Set("Origin", "https://example.test")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.test" {
		t.Fatal("missing CORS origin header")
	}
}

func TestCORSRejectsUnknownOrigin(t *testing.T) {
	handler := cors(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("disallowed preflight must not call next handler")
	}), []string{"https://allowed.test"})
	req := httptest.NewRequest(http.MethodOptions, "/api/events", nil)
	req.Header.Set("Origin", "https://attacker.test")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestWriteKeyProtectsMutations(t *testing.T) {
	called := false
	handler := requireWriteKey(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	}), "secret")

	unauthorized := httptest.NewRequest(http.MethodPost, "/api/events", nil)
	unauthorizedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedRecorder, unauthorized)
	if unauthorizedRecorder.Code != http.StatusUnauthorized || called {
		t.Fatalf("unauthorized mutation returned %d, called=%v", unauthorizedRecorder.Code, called)
	}

	authorized := httptest.NewRequest(http.MethodDelete, "/api/events/1", nil)
	authorized.Header.Set("Authorization", "Bearer secret")
	authorizedRecorder := httptest.NewRecorder()
	handler.ServeHTTP(authorizedRecorder, authorized)
	if authorizedRecorder.Code != http.StatusNoContent || !called {
		t.Fatalf("authorized mutation returned %d, called=%v", authorizedRecorder.Code, called)
	}
}
