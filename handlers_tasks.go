package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (app *App) handleAgentCreateTask(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	var req struct {
		Title       string  `json:"title"`
		Description string  `json:"description"`
		MaxBudget   float64 `json:"max_budget"`
		Currency    string  `json:"currency"`
		Deadline    string  `json:"deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Title == "" || req.Description == "" || req.MaxBudget <= 0 {
		writeError(w, http.StatusBadRequest, "title, description, and max_budget are required")
		return
	}
	if req.Currency == "" {
		req.Currency = "USDT"
	}
	if req.Currency != "USDT" && req.Currency != "DNZD" {
		writeError(w, http.StatusBadRequest, "currency must be USDT or DNZD")
		return
	}

	deadline, err := time.Parse(time.RFC3339, req.Deadline)
	if err != nil {
		writeError(w, http.StatusBadRequest, "deadline must be RFC3339 format")
		return
	}

	var taskID int64
	err = app.db.QueryRowContext(r.Context(),
		`INSERT INTO tasks (agent_id, title, description, max_budget, currency, deadline)
		 VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
		agentID, req.Title, req.Description, req.MaxBudget, req.Currency, deadline,
	).Scan(&taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	var t Task
	app.db.QueryRowContext(r.Context(),
		`SELECT id, agent_id, title, description, max_budget, currency, deadline, status, created_at
		 FROM tasks WHERE id = ?`, taskID,
	).Scan(&t.ID, &t.AgentID, &t.Title, &t.Description, &t.MaxBudget, &t.Currency, &t.Deadline, &t.Status, &t.CreatedAt)

	writeJSON(w, http.StatusCreated, t)
}

func (app *App) handleAgentListTasks(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	rows, err := app.db.QueryContext(r.Context(),
		`SELECT t.id, t.agent_id, t.title, t.description, t.max_budget, t.currency, t.deadline, t.status, t.created_at,
		        (SELECT COUNT(*) FROM bids WHERE task_id = t.id) as bid_count
		 FROM tasks t WHERE t.agent_id = ? ORDER BY t.created_at DESC`, agentID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()
	tasks := scanTasks(rows)
	writeJSON(w, http.StatusOK, tasks)
}

func (app *App) handleAgentGetTask(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var t Task
	err = app.db.QueryRowContext(r.Context(),
		`SELECT id, agent_id, title, description, max_budget, currency, deadline, status, created_at
		 FROM tasks WHERE id = ? AND agent_id = ?`, taskID, agentID,
	).Scan(&t.ID, &t.AgentID, &t.Title, &t.Description, &t.MaxBudget, &t.Currency, &t.Deadline, &t.Status, &t.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	bids, err := app.getBidsForTask(r, taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"task": t,
		"bids": bids,
	})
}

func (app *App) handlePublicListTasks(w http.ResponseWriter, r *http.Request) {
	currency := r.URL.Query().Get("currency")
	query := `SELECT t.id, t.agent_id, a.name as agent_name, t.title, t.description, t.max_budget, t.currency,
	                 t.deadline, t.status, t.created_at,
	                 (SELECT COUNT(*) FROM bids WHERE task_id = t.id) as bid_count
	          FROM tasks t JOIN agents a ON t.agent_id = a.id
	          WHERE t.status = 'open'`
	args := []interface{}{}
	if currency != "" {
		query += " AND t.currency = ?"
		args = append(args, currency)
	}
	query += " ORDER BY t.created_at DESC LIMIT 50"

	rows, err := app.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var t Task
		var agentName string
		rows.Scan(&t.ID, &t.AgentID, &agentName, &t.Title, &t.Description, &t.MaxBudget,
			&t.Currency, &t.Deadline, &t.Status, &t.CreatedAt, &t.BidCount)
		tasks = append(tasks, map[string]interface{}{
			"id": t.ID, "agent_id": t.AgentID, "agent_name": agentName,
			"title": t.Title, "description": t.Description, "max_budget": t.MaxBudget,
			"currency": t.Currency, "deadline": t.Deadline, "status": t.Status,
			"created_at": t.CreatedAt, "bid_count": t.BidCount,
		})
	}
	if tasks == nil {
		tasks = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (app *App) handlePublicGetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var t Task
	var agentName string
	err = app.db.QueryRowContext(r.Context(),
		`SELECT t.id, t.agent_id, a.name, t.title, t.description, t.max_budget, t.currency,
		        t.deadline, t.status, t.created_at
		 FROM tasks t JOIN agents a ON t.agent_id = a.id
		 WHERE t.id = ?`, taskID,
	).Scan(&t.ID, &t.AgentID, &agentName, &t.Title, &t.Description, &t.MaxBudget,
		&t.Currency, &t.Deadline, &t.Status, &t.CreatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	t.AgentName = agentName
	writeJSON(w, http.StatusOK, t)
}

func (app *App) handleAgentAcceptBid(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}
	bidID, err := strconv.ParseInt(chi.URLParam(r, "bid_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid bid id")
		return
	}

	// Verify agent owns this task
	var taskStatus string
	err = app.db.QueryRowContext(r.Context(),
		`SELECT status FROM tasks WHERE id = ? AND agent_id = ?`, taskID, agentID,
	).Scan(&taskStatus)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if taskStatus != "open" {
		writeError(w, http.StatusConflict, "task is not open")
		return
	}

	// Verify bid belongs to task and is pending
	var bidStatus string
	var bidAmount float64
	var bidCurrency string
	err = app.db.QueryRowContext(r.Context(),
		`SELECT b.status, b.amount, t.currency FROM bids b JOIN tasks t ON b.task_id = t.id
		 WHERE b.id = ? AND b.task_id = ?`, bidID, taskID,
	).Scan(&bidStatus, &bidAmount, &bidCurrency)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "bid not found")
		return
	}
	if bidStatus != "pending" {
		writeError(w, http.StatusConflict, "bid is not pending")
		return
	}

	tx, err := app.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer tx.Rollback()

	// Accept this bid, reject others
	if _, err := tx.ExecContext(r.Context(),
		`UPDATE bids SET status = 'accepted' WHERE id = ?`, bidID); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	if _, err := tx.ExecContext(r.Context(),
		`UPDATE bids SET status = 'rejected' WHERE task_id = ? AND id != ?`, taskID, bidID); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	// Move task to assigned
	if _, err := tx.ExecContext(r.Context(),
		`UPDATE tasks SET status = 'assigned' WHERE id = ?`, taskID); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	// Create escrow payment record
	if _, err := tx.ExecContext(r.Context(),
		`INSERT INTO payments (task_id, bid_id, amount, currency, status) VALUES (?, ?, ?, ?, 'escrow')`,
		taskID, bidID, bidAmount, bidCurrency); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}

func (app *App) handleAgentReviewDelivery(w http.ResponseWriter, r *http.Request) {
	agentID := getAgentID(r)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req struct {
		Action  string `json:"action"` // "approve" or "reject"
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Action != "approve" && req.Action != "reject" {
		writeError(w, http.StatusBadRequest, "action must be 'approve' or 'reject'")
		return
	}

	var taskStatus string
	err = app.db.QueryRowContext(r.Context(),
		`SELECT status FROM tasks WHERE id = ? AND agent_id = ?`, taskID, agentID,
	).Scan(&taskStatus)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if taskStatus != "delivered" {
		writeError(w, http.StatusConflict, "task has no pending delivery")
		return
	}

	tx, err := app.db.BeginTx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer tx.Rollback()

	now := time.Now()
	if req.Action == "approve" {
		tx.ExecContext(r.Context(),
			`UPDATE deliveries SET status = 'approved', reviewed_at = ? WHERE task_id = ? AND status = 'pending'`,
			now, taskID)
		tx.ExecContext(r.Context(),
			`UPDATE tasks SET status = 'completed' WHERE id = ?`, taskID)
		tx.ExecContext(r.Context(),
			`UPDATE payments SET status = 'released' WHERE task_id = ?`, taskID)
	} else {
		tx.ExecContext(r.Context(),
			`UPDATE deliveries SET status = 'rejected', reviewed_at = ? WHERE task_id = ? AND status = 'pending'`,
			now, taskID)
		tx.ExecContext(r.Context(),
			`UPDATE tasks SET status = 'disputed' WHERE id = ?`, taskID)
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": req.Action + "d"})
}

func (app *App) getBidsForTask(r *http.Request, taskID int64) ([]Bid, error) {
	rows, err := app.db.QueryContext(r.Context(),
		`SELECT b.id, b.task_id, b.user_id, u.display_name, b.amount, b.delivery_days, b.message, b.status, b.created_at
		 FROM bids b JOIN users u ON b.user_id = u.id
		 WHERE b.task_id = ? ORDER BY b.amount ASC`, taskID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bids []Bid
	for rows.Next() {
		var b Bid
		rows.Scan(&b.ID, &b.TaskID, &b.UserID, &b.DisplayName, &b.Amount, &b.DeliveryDays, &b.Message, &b.Status, &b.CreatedAt)
		bids = append(bids, b)
	}
	if bids == nil {
		bids = []Bid{}
	}
	return bids, nil
}

func scanTasks(rows *sql.Rows) []Task {
	var tasks []Task
	for rows.Next() {
		var t Task
		rows.Scan(&t.ID, &t.AgentID, &t.Title, &t.Description, &t.MaxBudget, &t.Currency, &t.Deadline, &t.Status, &t.CreatedAt, &t.BidCount)
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks
}
