package middleware

import (
	"context"
	"flight-booking/internal/service"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	tokenService service.TokenService
}

func NewAuthMiddleware(tokenService service.TokenService) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

func SetUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDKey, userID)
}

type contextKey string

const userIDKey contextKey = "userID"

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		if token == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		claims, err := m.tokenService.Verify(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		if claims == nil || claims.UserID == "" {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
