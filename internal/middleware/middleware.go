package middleware

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/romeulima/devbook-server/internal/security"
)

func Authentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		token, err := security.ValidateToken(header)

		if err != nil {
			slog.Error("error during token validation", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userId := chi.URLParam(r, "id")

		sub := token.Claims.(jwt.MapClaims)["sub"]

		if userId != sub {
			slog.Error("different users", "error", "user id is not same than token subject")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
