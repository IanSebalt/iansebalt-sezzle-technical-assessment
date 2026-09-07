package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/iansebalt/sezzle-calculator/internal/apierror"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (a *api) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// The status line is already sent, so this can only be logged.
		a.logger.Error("failed to encode response body", "error", err)
	}
}

func (a *api) writeError(w http.ResponseWriter, apiErr *apierror.Error) {
	if apiErr.Status >= http.StatusInternalServerError {
		a.logger.Error("request failed", "error", apiErr, "status", apiErr.Status)
	}

	a.writeJSON(w, apiErr.Status, errorBody{
		Error: errorDetail{Code: apiErr.Code, Message: apiErr.Message},
	})
}
