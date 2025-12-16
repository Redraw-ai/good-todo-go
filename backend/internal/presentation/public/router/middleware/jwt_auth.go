package middleware

import (
	"net/http"
	"strings"

	"good-todo-go/internal/infrastructure/database"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/presentation/public/router/context_keys"

	"github.com/labstack/echo/v4"
)

// 認証が不要なルートの一覧
var publicRoutes = []string{
	"/health",
	"/auth/register",
	"/auth/login",
	"/auth/verify-email",
	"/auth/refresh",
}

// JWTAuthMiddleware validates JWT tokens and sets user info in context
func JWTAuthMiddleware(jwtService *pkg.JWTService) echo.MiddlewareFunc {
	// publicRoutesをmapに変換
	publicRoutesMap := make(map[string]bool)
	for _, route := range publicRoutes {
		publicRoutesMap[route] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// CORSのプリフライトリクエストに対する処理
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}

			// 認証不要なルートの場合は、認証をスキップする
			if publicRoutesMap[c.Request().URL.Path] {
				return next(c)
			}

			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Remove "Bearer " prefix
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			// Validate token
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}

			if claims.TokenType != pkg.AccessToken {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token type")
			}

			// Set user info in echo context
			c.Set(context_keys.UserIDContextKey, claims.UserID)
			c.Set(context_keys.TenantIDContextKey, claims.TenantID)
			c.Set(context_keys.EmailContextKey, claims.Email)
			c.Set(context_keys.RoleContextKey, claims.Role)

			// Also set tenantID in request context for database layer
			ctx := database.WithTenantID(c.Request().Context(), claims.TenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// JWTAuth は後方互換性のためのエイリアス
func JWTAuth(jwtService *pkg.JWTService) echo.MiddlewareFunc {
	return JWTAuthMiddleware(jwtService)
}
