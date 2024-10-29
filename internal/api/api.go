package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	m "github.com/romeulima/devbook-server/internal/middleware"
	"github.com/romeulima/devbook-server/internal/models"
	"github.com/romeulima/devbook-server/internal/security"
	"github.com/romeulima/devbook-server/internal/storage"
	"github.com/romeulima/devbook-server/pkg"
	"golang.org/x/crypto/bcrypt"
)

type handler struct {
	q storage.PgStore
}

func NewHandler(db storage.PgStore) http.Handler {

	h := handler{
		q: db,
	}

	r := chi.NewMux()

	r.Use(middleware.Logger, middleware.Recoverer, middleware.RequestID)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", h.handleCreateUser)
		r.Get("/", h.handleGetUsers)
		r.Get("/{id}", h.handleGetUserByID)
		r.Put("/{id}", h.handleUpdateUser)
		r.Group(func(r chi.Router) {
			r.Delete("/{id}", m.Authentication(h.handleDeleteUser))
		})
	})

	r.Post("/login", h.handleLogin)

	return r
}

func (h *handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginPayload models.UserPayload

	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&loginPayload); err != nil {
		pkg.SendJSON(w, models.Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
		return
	}

	userId, password, err := h.q.GetUserByEmail(r.Context(), loginPayload.Email)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			pkg.SendJSON(w, models.Response{Error: "user not found"}, http.StatusNotFound)
			return
		}
		slog.Error("error to search user by email", "error", err)
		pkg.SendJSON(w, models.Response{Error: "something wetn wrong"}, http.StatusInternalServerError)
		return
	}

	if err = security.ValidatePasswords(password, loginPayload.Password); err != nil {
		pkg.SendJSON(w, models.Response{Error: "email or password invalids"}, http.StatusUnauthorized)
		return
	}

	token, err := security.GenerateToken(userId.String())

	if err != nil {
		slog.Error("error to generate token", "error", err)
		pkg.SendJSON(w, models.Response{Error: "something wetn wrong"}, http.StatusInternalServerError)
		return
	}

	pkg.SendJSON(w, models.Response{Data: token}, http.StatusOK)
}

func (h *handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userPayload models.UserPayload
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		pkg.SendJSON(w, models.Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
		return
	}

	if err := userPayload.PrepareUser(&userPayload); err != nil {
		var missingFields *models.MissingFieldError
		if errors.As(err, &missingFields) {
			pkg.SendJSON(w, models.Response{Error: missingFields.Error()}, http.StatusBadRequest)
			return
		}
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			pkg.SendJSON(w, models.Response{Error: "password is bigger than requested"}, http.StatusBadRequest)
			return
		}

		slog.Error("error when preparing user", "error", err)
		pkg.SendJSON(w, models.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	userResponse, err := h.q.CreateUser(ctx, userPayload)
	if err != nil {
		slog.Error("Failed to insert user", "error", err, "email", userPayload.Email)
		pkg.SendJSON(w, models.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}
	pkg.SendJSON(w, models.Response{Data: userResponse}, http.StatusCreated)
}

func (h *handler) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	param := strings.ToLower(r.URL.Query().Get("user"))

	users, err := h.q.GetUsers(r.Context(), param)

	if err != nil {
		slog.Error("failed to get users by query params", "error", err)
		pkg.SendJSON(w, models.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}

	if users == nil {
		pkg.SendJSON(w, models.Response{Data: []models.UserResponse{}}, http.StatusOK)
		return
	}

	pkg.SendJSON(w, models.Response{Data: users}, http.StatusOK)
}

func (h *handler) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "id")

	id, err := uuid.Parse(param)

	if err != nil {
		pkg.SendJSON(w, models.Response{Error: "invalid uuid"}, http.StatusBadRequest)
		return
	}

	user, err := h.q.GetUserByID(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			pkg.SendJSON(w, models.Response{Error: "user not found"}, http.StatusNotFound)
			return
		}
		slog.Error("Failed to get user by id", "error", err)
		pkg.SendJSON(w, models.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}
	pkg.SendJSON(w, models.Response{Data: user}, http.StatusOK)
}

func (h *handler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "id")

	id := pkg.VerifyUUID(w, param)

	if id == uuid.Nil {
		return
	}

	var data models.UserPayload
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		pkg.SendJSON(w, models.Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
		return
	}

	if err := h.q.UpdateUser(r.Context(), id, data); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			pkg.SendJSON(w, models.Response{Error: "user not found"}, http.StatusNotFound)
			return
		}
		slog.Error("error to update user", "error", err)
		pkg.SendJSON(w, models.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}
	pkg.SendJSON(w, models.Response{Data: nil}, http.StatusNoContent)
}

func (h *handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	param := chi.URLParam(r, "id")

	id := pkg.VerifyUUID(w, param)

	if id == uuid.Nil {
		return
	}

	if err := h.q.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			pkg.SendJSON(w, models.Response{Error: "user not found"}, http.StatusNotFound)
			return
		}
		slog.Error("error to delete user", "error", err)
		pkg.SendJSON(w, models.Response{Error: "something went wrong"}, http.StatusInternalServerError)
		return
	}
	pkg.SendJSON(w, models.Response{Data: nil}, http.StatusNoContent)
}
