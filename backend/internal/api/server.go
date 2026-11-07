package api

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/n1ckerr0r/chronograph-spb/backend/internal/openapi"
)

// dateLayout задает публичный формат дат для событий и фильтров API.
const dateLayout = "2006-01-02"

// API хранит доступ к PostgreSQL и регистрирует HTTP-обработчики.
type API struct {
	db *pgxpool.Pool
}

// New создает HTTP-роутер для Wandero API.
func New(db *pgxpool.Pool, allowedOrigins []string, writeKey string) http.Handler {
	api := &API{db: db}

	handler := openapi.HandlerWithOptions(api, openapi.StdHTTPServerOptions{
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, http.StatusBadRequest, err.Error())
		},
	})

	return recoverPanic(securityHeaders(cors(requireWriteKey(handler, writeKey), allowedOrigins)))
}

// health проверяет доступность HTTP-сервера и подключения к базе данных.
func (api *API) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := api.db.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database is unavailable")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
