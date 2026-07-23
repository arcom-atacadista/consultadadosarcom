package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/arcom-atacadista/consultadadosarcom/backend/internal/httputil"
)

type ctxKey string

const ctxClaims ctxKey = "auth.claims"

// FromContext devolve as claims do usuário autenticado (só depois de RequireAuth).
func FromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(ctxClaims).(*Claims)
	return c, ok
}

// RequireAuth exige um Bearer token válido e injeta as claims no contexto.
func RequireAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				httputil.WriteError(w, http.StatusUnauthorized, "faça login para continuar")
				return
			}
			claims, err := parseToken(jwtSecret, token)
			if err != nil {
				httputil.WriteError(w, http.StatusUnauthorized, "sessão inválida ou expirada")
				return
			}
			ctx := context.WithValue(r.Context(), ctxClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAprovado exige que o usuário (já autenticado) esteja aprovado ou seja admin.
func RequireAprovado(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := FromContext(r.Context())
		if !ok || !claims.Aprovado {
			httputil.WriteError(w, http.StatusForbidden, "sua conta ainda não foi aprovada")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin exige que o usuário (já autenticado) seja administrador.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := FromContext(r.Context())
		if !ok || !claims.IsAdmin {
			httputil.WriteError(w, http.StatusForbidden, "ação restrita a administradores")
			return
		}
		next.ServeHTTP(w, r)
	})
}
