package api

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func replaceEventTags(ctx context.Context, tx pgx.Tx, eventID int64, tags []string) error {
	if _, err := tx.Exec(ctx, `DELETE FROM event_tag WHERE event_id=$1`, eventID); err != nil {
		return err
	}

	seen := make(map[string]struct{})
	for _, raw := range tags {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}

		var tagID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO tag (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name=EXCLUDED.name
			RETURNING id`, name,
		).Scan(&tagID)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO event_tag (event_id, tag_id) VALUES ($1, $2)`, eventID, tagID); err != nil {
			return err
		}
	}

	return nil
}

func scanSources(rows pgx.Rows) ([]sourceDTO, error) {
	items := make([]sourceDTO, 0)
	for rows.Next() {
		var item sourceDTO
		var year sql.NullInt64
		if err := rows.Scan(&item.ID, &item.Title, &item.Author, &year, &item.URL, &item.Type); err != nil {
			return nil, err
		}
		if year.Valid {
			value := int(year.Int64)
			item.Year = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func parseFlexibleDate(value string, endOfYear bool) (time.Time, error) {
	if len(value) == 4 {
		year, err := strconv.Atoi(value)
		if err != nil {
			return time.Time{}, badRequestError("date filter must use YYYY or YYYY-MM-DD")
		}
		if endOfYear {
			return time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC), nil
		}
		return time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC), nil
	}

	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, badRequestError("date filter must use YYYY or YYYY-MM-DD")
	}
	return parsed, nil
}

func parsePositiveInt(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0, badRequestError("pagination values must be non-negative integers")
	}
	if parsed > 500 {
		return 0, badRequestError("pagination limit must not exceed 500")
	}
	return parsed, nil
}

// parseOffset limits deep pagination independently from the smaller page-size limit.
func parseOffset(value string, fallback int) (int, error) {
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || parsed > 100_000 {
		return 0, badRequestError("pagination offset must be between 0 and 100000")
	}
	return parsed, nil
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	value := r.PathValue(name)
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, name+" must be a positive integer")
		return 0, false
	}
	return id, true
}

func readJSONBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "request body must contain exactly one JSON object")
		return false
	}
	return true
}

func handleDBError(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503":
			writeError(w, http.StatusBadRequest, "referenced resource does not exist")
			return
		case "23505":
			writeError(w, http.StatusConflict, "resource already exists")
			return
		case "23514", "22P02":
			writeError(w, http.StatusBadRequest, "request violates a data constraint")
			return
		}
	}
	log.Printf("database error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

type badRequestError string

func (err badRequestError) Error() string {
	return string(err)
}

func handleAPIError(w http.ResponseWriter, err error) {
	var badRequest badRequestError
	if errors.As(err, &badRequest) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	handleDBError(w, err)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func cors(next http.Handler, allowedOrigins []string) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		_, isAllowed := allowed[origin]
		if origin != "" && isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			if origin != "" && !isAllowed {
				writeError(w, http.StatusForbidden, "origin is not allowed")
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("panic serving %s %s: %v", r.Method, r.URL.Path, recovered)
				writeError(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func requireWriteKey(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key == "" || (r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete) {
			next.ServeHTTP(w, r)
			return
		}
		const prefix = "Bearer "
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, prefix) || subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(authorization, prefix)), []byte(key)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
