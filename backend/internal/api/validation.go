package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func readLocationInput(w http.ResponseWriter, r *http.Request) (locationInput, bool) {
	var input locationInput
	if !readJSONBody(w, r, &input) {
		return locationInput{}, false
	}
	if strings.TrimSpace(input.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return locationInput{}, false
	}
	if input.Latitude < -90 || input.Latitude > 90 {
		writeError(w, http.StatusBadRequest, "latitude must be between -90 and 90")
		return locationInput{}, false
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		writeError(w, http.StatusBadRequest, "longitude must be between -180 and 180")
		return locationInput{}, false
	}
	return input, true
}

func readEventInput(w http.ResponseWriter, r *http.Request) (eventInput, bool) {
	var input eventInput
	if !readJSONBody(w, r, &input) {
		return eventInput{}, false
	}
	if strings.TrimSpace(input.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return eventInput{}, false
	}
	if _, _, err := parseEventDates(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return eventInput{}, false
	}
	return input, true
}

func readRouteInput(w http.ResponseWriter, r *http.Request) (routeInput, bool) {
	var input routeInput
	if !readJSONBody(w, r, &input) {
		return routeInput{}, false
	}
	if strings.TrimSpace(input.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return routeInput{}, false
	}
	if len(input.Stops) == 0 || len(input.Stops) > 12 {
		writeError(w, http.StatusBadRequest, "stops must contain between 1 and 12 places")
		return routeInput{}, false
	}
	seen := make(map[int64]struct{}, len(input.Stops))
	for _, stop := range input.Stops {
		if stop.LocationID <= 0 {
			writeError(w, http.StatusBadRequest, "every stop must have a positive location_id")
			return routeInput{}, false
		}
		if stop.VisitMinutes < 0 || stop.VisitMinutes > 1440 {
			writeError(w, http.StatusBadRequest, "visit_minutes must be between 0 and 1440")
			return routeInput{}, false
		}
		if _, duplicate := seen[stop.LocationID]; duplicate {
			writeError(w, http.StatusBadRequest, "a location can only appear once in a route")
			return routeInput{}, false
		}
		seen[stop.LocationID] = struct{}{}
	}
	for _, value := range []*int{input.DistanceMeters, input.WalkingMinutes} {
		if value != nil && *value < 0 {
			writeError(w, http.StatusBadRequest, "route durations and distances cannot be negative")
			return routeInput{}, false
		}
	}
	return input, true
}

func readSourceInput(w http.ResponseWriter, r *http.Request) (sourceInput, bool) {
	var input sourceInput
	if !readJSONBody(w, r, &input) {
		return sourceInput{}, false
	}
	if strings.TrimSpace(input.Title) == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return sourceInput{}, false
	}
	if strings.TrimSpace(input.Type) == "" {
		input.Type = "web"
	}
	if !isSourceType(input.Type) {
		writeError(w, http.StatusBadRequest, "type must be one of book, article, archive, photo, map, web")
		return sourceInput{}, false
	}
	if input.Year != nil && (*input.Year < 0 || *input.Year > time.Now().Year()+1) {
		writeError(w, http.StatusBadRequest, "year is out of range")
		return sourceInput{}, false
	}
	if strings.TrimSpace(input.URL) != "" {
		parsed, err := url.ParseRequestURI(input.URL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			writeError(w, http.StatusBadRequest, "url must be an absolute HTTP(S) URL")
			return sourceInput{}, false
		}
	}
	return input, true
}

func parseEventDates(input eventInput) (time.Time, *time.Time, error) {
	dateFrom, err := time.Parse(dateLayout, input.DateFrom)
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("date_from must use YYYY-MM-DD")
	}

	var dateTo *time.Time
	if input.DateTo != nil && strings.TrimSpace(*input.DateTo) != "" {
		parsed, err := time.Parse(dateLayout, *input.DateTo)
		if err != nil {
			return time.Time{}, nil, fmt.Errorf("date_to must use YYYY-MM-DD")
		}
		if parsed.Before(dateFrom) {
			return time.Time{}, nil, fmt.Errorf("date_to must be greater than or equal to date_from")
		}
		dateTo = &parsed
	}

	return dateFrom, dateTo, nil
}

func isSourceType(value string) bool {
	switch value {
	case "book", "article", "archive", "photo", "map", "web":
		return true
	default:
		return false
	}
}
