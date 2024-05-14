package jwt

import (
	"banner-serivce/internal/api/response"
	"github.com/go-chi/render"
	"net/http"
	"strings"
)

func TokenAuthMiddleware(jwtManager *JWTManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			render.JSON(w, r, response.Error("unauthorized"))
			return
		}

		token := strings.Split(tokenString, " ")
		if len(token) != 2 || token[0] != "Bearer" {
			render.JSON(w, r, response.Error("unauthorized"))
			return
		}

		_, err := jwtManager.VerifyToken(token[1])
		if err != nil {
			render.JSON(w, r, response.Error("unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TokenAuthAndRoleMiddleware(jwtManager *JWTManager, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			render.JSON(w, r, response.Error("unauthorized"))
			return
		}

		token := strings.Split(tokenString, " ")
		if len(token) != 2 || token[0] != "Bearer" {
			render.JSON(w, r, response.Error("unauthorized"))
			return
		}

		claims, err := jwtManager.VerifyToken(token[1])
		if err != nil {
			render.JSON(w, r, response.Error("Unauthorized"))
			return
		}

		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			render.JSON(w, r, response.Error("forbidden"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
