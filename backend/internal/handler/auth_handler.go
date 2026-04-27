package handler

import (
	"encoding/json"
	"net/http"

	"github.com/fenmo/expense-tracker/internal/model"
	"github.com/fenmo/expense-tracker/internal/service"
	"go.uber.org/zap"
)

type AuthHandler struct {
	svc    service.AuthService
	logger *zap.Logger
}

func NewAuthHandler(svc service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input model.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errResponse{Code: "INVALID_JSON", Message: "request body must be valid JSON"})
		return
	}
	resp, err := h.svc.Register(r.Context(), input)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input model.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errResponse{Code: "INVALID_JSON", Message: "request body must be valid JSON"})
		return
	}
	resp, err := h.svc.Login(r.Context(), input)
	if err != nil {
		writeError(w, h.logger, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
