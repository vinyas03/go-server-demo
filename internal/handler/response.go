package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go-learn/internal/model"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// JSON writes any data as a JSON payload with the specified status code.
func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		}
	}
}

// Error writes a structured JSON error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, ErrorResponse{Error: message})
}

// RespondWithError maps domain errors to appropriate HTTP status codes.
func RespondWithError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrNotFound):
		Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, model.ErrInvalidInput):
		Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, model.ErrAlreadyExists):
		Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, model.ErrUnauthorized):
		Error(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, model.ErrForbidden):
		Error(w, http.StatusForbidden, err.Error())
	default:
		Error(w, http.StatusInternalServerError, "internal server error")
	}
}
