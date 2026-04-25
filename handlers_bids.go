package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func (app *App) handlePublicListBids(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	bids, err := app.getBidsForTask(r, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, bids)
}

func (app *App) handleSubmitBid(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req struct {
		Amount       float64 `json:"amount"`
		DeliveryDays int     `json:"delivery_days"`
		Message      string  `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Amount <= 0 || req.DeliveryDays <= 0 {
		writeError(w, http.StatusBadRequest, "amount and delivery_days must be positive")
		return
	}

	// Verify task is open and get max budget
	var taskStatus string
	var maxBudget float64
	err = app.db.QueryRowContext(r.Context(),
		`SELECT status, max_budget FROM tasks WHERE id = ?`, taskID,
	).Scan(&taskStatus, &maxBudget)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if taskStatus != "open" {
		writeError(w, http.StatusConflict, "task is not accepting bids")
		return
	}
	if req.Amount > maxBudget {
		writeError(w, http.StatusBadRequest, "bid amount exceeds task max budget")
		return
	}

	// Check if user already has a pending bid on this task
	var existingCount int
	app.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM bids WHERE task_id = ? AND user_id = ? AND status = 'pending'`,
		taskID, userID,
	).Scan(&existingCount)
	if existingCount > 0 {
		writeError(w, http.StatusConflict, "you already have a pending bid on this task")
		return
	}

	var bidID int64
	err = app.db.QueryRowContext(r.Context(),
		`INSERT INTO bids (task_id, user_id, amount, delivery_days, message) VALUES (?, ?, ?, ?, ?) RETURNING id`,
		taskID, userID, req.Amount, req.DeliveryDays, req.Message,
	).Scan(&bidID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create bid")
		return
	}

	var b Bid
	app.db.QueryRowContext(r.Context(),
		`SELECT id, task_id, user_id, amount, delivery_days, message, status, created_at
		 FROM bids WHERE id = ?`, bidID,
	).Scan(&b.ID, &b.TaskID, &b.UserID, &b.Amount, &b.DeliveryDays, &b.Message, &b.Status, &b.CreatedAt)

	writeJSON(w, http.StatusCreated, b)
}

func (app *App) handleMyBids(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	rows, err := app.db.QueryContext(r.Context(),
		`SELECT b.id, b.task_id, b.user_id, b.amount, b.delivery_days, b.message, b.status, b.created_at,
		        t.title, t.currency, t.status as task_status
		 FROM bids b JOIN tasks t ON b.task_id = t.id
		 WHERE b.user_id = ? ORDER BY b.created_at DESC`, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()

	type BidWithTask struct {
		Bid
		TaskTitle    string `json:"task_title"`
		TaskCurrency string `json:"task_currency"`
		TaskStatus   string `json:"task_status"`
	}
	var bids []BidWithTask
	for rows.Next() {
		var b BidWithTask
		rows.Scan(&b.ID, &b.TaskID, &b.UserID, &b.Amount, &b.DeliveryDays, &b.Message, &b.Status, &b.CreatedAt,
			&b.TaskTitle, &b.TaskCurrency, &b.TaskStatus)
		bids = append(bids, b)
	}
	if bids == nil {
		bids = []BidWithTask{}
	}
	writeJSON(w, http.StatusOK, bids)
}

func (app *App) handleWithdrawBid(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	bidID, err := strconv.ParseInt(chi.URLParam(r, "bid_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bid id")
		return
	}

	result, err := app.db.ExecContext(r.Context(),
		`UPDATE bids SET status = 'withdrawn' WHERE id = ? AND user_id = ? AND status = 'pending'`,
		bidID, userID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "bid not found or not withdrawable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "withdrawn"})
}
