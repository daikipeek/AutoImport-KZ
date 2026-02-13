package middleware

import (
	"context"
	"net/http"
	"strconv"

	"autoimport-kz/internal/models"
	"autoimport-kz/internal/repositories"
	"github.com/gorilla/sessions"
)

type contextKey string

const userKey contextKey = "user"

func UserFromContext(ctx context.Context) (models.User, bool) {
	u, ok := ctx.Value(userKey).(models.User)
	return u, ok
}

func AuthMiddleware(store *sessions.CookieStore, sessionName string, users *repositories.UserRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := store.Get(r, sessionName)
			idVal, ok := session.Values["user_id"]
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			id, ok := sessionIDToInt64(idVal)
			if !ok {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			user, err := users.ByID(r.Context(), id)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func sessionIDToInt64(v any) (int64, bool) {
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case int32:
		return int64(id), true
	case float64:
		return int64(id), true
	case string:
		parsed, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func RoleMiddleware(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || user.Role != role {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
