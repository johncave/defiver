package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func generateAPIKey() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	key := "ak_" + hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(key))
	return key, hex.EncodeToString(hash[:]), nil
}

func (app *App) handleAgentRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name               string `json:"name"`
		Email              string `json:"email"`
		CurrencyPreference string `json:"currency_preference"`
		WalletAddress      string `json:"wallet_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Email == "" {
		writeError(w, http.StatusBadRequest, "name and email are required")
		return
	}
	if req.CurrencyPreference == "" {
		req.CurrencyPreference = "USDT"
	}

	apiKey, apiKeyHash, err := generateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate API key")
		return
	}

	var agentID int64
	err = app.db.QueryRowContext(r.Context(),
		`INSERT INTO agents (api_key_hash, name, email, currency_preference, wallet_address)
		 VALUES (?, ?, ?, ?, ?) RETURNING id`,
		apiKeyHash, req.Name, req.Email, req.CurrencyPreference, req.WalletAddress,
	).Scan(&agentID)
	if err != nil {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"agent_id": agentID,
		"api_key":  apiKey,
		"message":  "Store your API key securely — it will not be shown again.",
	})
}

func (app *App) handleAgentMe(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	var a Agent
	err := app.db.QueryRowContext(r.Context(),
		`SELECT id, name, email, currency_preference, wallet_address, created_at
		 FROM agents WHERE id = ?`, agentID,
	).Scan(&a.ID, &a.Name, &a.Email, &a.CurrencyPreference, &a.WalletAddress, &a.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (app *App) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "email, password, and display_name are required")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	var userID int64
	err = app.db.QueryRowContext(r.Context(),
		`INSERT INTO users (email, password_hash, display_name) VALUES (?, ?, ?) RETURNING id`,
		req.Email, string(hash), req.DisplayName,
	).Scan(&userID)
	if err != nil {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}

	token, err := app.signToken(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign token")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"token":        token,
		"user_id":      userID,
		"display_name": req.DisplayName,
	})
}

func (app *App) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var u User
	err := app.db.QueryRowContext(r.Context(),
		`SELECT id, email, password_hash, display_name FROM users WHERE email = ?`, req.Email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := app.signToken(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to sign token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":        token,
		"user_id":      u.ID,
		"display_name": u.DisplayName,
	})
}

func (app *App) handleUserMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	var u User
	err := app.db.QueryRowContext(r.Context(),
		`SELECT id, email, display_name, wallet_address, bio, reputation_score, created_at
		 FROM users WHERE id = ?`, userID,
	).Scan(&u.ID, &u.Email, &u.DisplayName, &u.WalletAddress, &u.Bio, &u.ReputationScore, &u.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (app *App) signToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(app.jwtSecret))
}
