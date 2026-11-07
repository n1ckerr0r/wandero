package api

import "time"

type locationDTO struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type locationInput struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

type eventDTO struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	DateFrom     string    `json:"date_from"`
	DateTo       *string   `json:"date_to,omitempty"`
	LocationID   *int64    `json:"location_id,omitempty"`
	LocationName string    `json:"location_name,omitempty"`
	Latitude     *float64  `json:"latitude,omitempty"`
	Longitude    *float64  `json:"longitude,omitempty"`
	Tags         []string  `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type eventInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	DateFrom    string   `json:"date_from"`
	DateTo      *string  `json:"date_to"`
	LocationID  *int64   `json:"location_id"`
	Tags        []string `json:"tags"`
}

type tagDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type tagInput struct {
	Name string `json:"name"`
}

type sourceDTO struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   *int   `json:"year,omitempty"`
	URL    string `json:"url"`
	Type   string `json:"type"`
}

type sourceInput struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   *int   `json:"year"`
	URL    string `json:"url"`
	Type   string `json:"type"`
}

type mediaDTO struct {
	ID      int64  `json:"id"`
	EventID int64  `json:"event_id"`
	URL     string `json:"url"`
	Caption string `json:"caption"`
	Type    string `json:"type"`
}

type mediaInput struct {
	URL     string `json:"url"`
	Caption string `json:"caption"`
	Type    string `json:"type"`
}

type routeStopInput struct {
	LocationID   int64 `json:"location_id"`
	VisitMinutes int   `json:"visit_minutes"`
}

type routeInput struct {
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Published      bool             `json:"published"`
	DistanceMeters *int             `json:"distance_meters"`
	WalkingMinutes *int             `json:"walking_minutes"`
	Stops          []routeStopInput `json:"stops"`
}

type routeStopDTO struct {
	Position     int         `json:"position"`
	LocationID   int64       `json:"location_id"`
	VisitMinutes int         `json:"visit_minutes"`
	Location     locationDTO `json:"location"`
}

type routeDTO struct {
	ID              int64          `json:"id"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Published       bool           `json:"published"`
	DistanceMeters  *int           `json:"distance_meters"`
	WalkingMinutes  *int           `json:"walking_minutes"`
	CompletionCount int            `json:"completion_count"`
	Stops           []routeStopDTO `json:"stops"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
