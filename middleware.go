package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ctxUserID  contextKey = "userID"
	ctxAgentID contextKey = "agentID"
)

func (app *App) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(app.jwtSecret), nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}
		userID, ok := claims["sub"].(float64)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid token subject")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, int64(userID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *App) requireAgent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			writeError(w, http.StatusUnauthorized, "missing X-API-Key")
			return
		}
		hash := sha256.Sum256([]byte(apiKey))
		hashStr := hex.EncodeToString(hash[:])

		var agentID int64
		err := app.db.QueryRowContext(r.Context(),
			`SELECT id FROM agents WHERE api_key_hash = ?`, hashStr,
		).Scan(&agentID)
		if err == sql.ErrNoRows {
			writeError(w, http.StatusUnauthorized, "invalid API key")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "db error")
			return
		}
		ctx := context.WithValue(r.Context(), ctxAgentID, agentID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getUserID(r *http.Request) int64 {
	id, _ := r.Context().Value(ctxUserID).(int64)
	return id
}

func getAgentID(r *http.Request) int64 {
	id, _ := r.Context().Value(ctxAgentID).(int64)
	return id
}
