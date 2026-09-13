package httperr

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/rs/zerolog"
)

var (
	ErrInvalid   = errors.New("invalid_request")
	ErrNotFound  = errors.New("not_found")
	ErrForbidden = errors.New("forbidden")
	ErrConflict  = errors.New("conflict")
)

func Write(w http.ResponseWriter, log zerolog.Logger, err error) {
	switch {
	case errors.Is(err, ErrInvalid), errors.Is(err, auth.ErrInvalidRequest):
		utils.JSONError(w, http.StatusBadRequest, "invalid_request", "Check the request and try again.")
	case errors.Is(err, ErrNotFound), errors.Is(err, pgx.ErrNoRows):
		utils.JSONError(w, http.StatusNotFound, "not_found", "Resource not found.")
	case errors.Is(err, ErrForbidden):
		utils.JSONError(w, http.StatusForbidden, "forbidden", "You do not have access.")
	case errors.Is(err, ErrConflict), isUniqueViolation(err):
		utils.JSONError(w, http.StatusConflict, "conflict", "Resource already exists.")
	case errors.Is(err, auth.ErrUnauthenticated):
		utils.JSONError(w, http.StatusUnauthorized, "unauthenticated", "Sign in required.")
	default:
		var maxBytes *http.MaxBytesError
		if errors.As(err, &maxBytes) {
			utils.JSONError(w, http.StatusRequestEntityTooLarge, "invalid_request", "Request is too large.")
			return
		}
		log.Error().Err(err).Msg("handler")
		utils.JSONError(w, http.StatusInternalServerError, "internal", "Something went wrong.")
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
