package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fenmo/expense-tracker/internal/model"
	"go.uber.org/zap"
)

type errResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func writeError(w http.ResponseWriter, logger *zap.Logger, err error) {
	var status int
	var code, msg string

	switch {
	case errors.Is(err, model.ErrInvalidInput):
		status, code = http.StatusBadRequest, "INVALID_INPUT"
		msg = err.Error()
	case errors.Is(err, model.ErrDuplicateKey):
		status, code = http.StatusConflict, "IDEMPOTENCY_CONFLICT"
		msg = err.Error()
	case errors.Is(err, model.ErrNotFound):
		status, code = http.StatusNotFound, "NOT_FOUND"
		msg = err.Error()
	default:
		status, code = http.StatusInternalServerError, "INTERNAL_ERROR"
		msg = "an unexpected error occurred"
		logger.Error("unhandled error", zap.Error(err))
	}

	writeJSON(w, status, errResponse{Code: code, Message: msg})
}
