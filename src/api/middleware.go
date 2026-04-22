package main

import (
	"context"
	"log"
	"net/http"
)

func NewLoggerMiddleware(logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("New request with method: %s , path: %s , remote address: %s", r.Method, r.URL.Path, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

func NewAuthMiddleware(jwtManager *JWTManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.URL.Path == healthCheckRoute || r.URL.Path == signUpRoute || r.URL.Path == signInRoute {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[len("Bearer "):]

			isValid, err := jwtManager.IsAccessToken(tokenString)
			if err != nil || !isValid {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			userId, err := jwtManager.GetUserIDFromToken(tokenString)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Store the user ID in the request context for later retrieval in handlers
			ctx := r.Context()
			ctx = context.WithValue(ctx, "userID", userId)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
