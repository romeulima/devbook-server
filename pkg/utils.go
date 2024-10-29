package pkg

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/romeulima/devbook-server/internal/models"
)

func SendJSON(w http.ResponseWriter, resp models.Response, status int) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(resp)

	if err != nil {
		slog.Error("failed to marshal json data", "error", err)
		SendJSON(w, models.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("failed to write response to client", "error", err)
		return
	}
}

func VerifyUUID(w http.ResponseWriter, param string) uuid.UUID {
	id, err := uuid.Parse(param)

	if err != nil {
		SendJSON(w, models.Response{Error: "invalid uuid"}, http.StatusBadRequest)
		return uuid.Nil
	}

	return id
}
