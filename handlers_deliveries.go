package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func (app *App) handleSubmitDelivery(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	// Verify user has an accepted bid on this task
	var bidID int64
	err = app.db.QueryRowContext(r.Context(),
		`SELECT id FROM bids WHERE task_id = ? AND user_id = ? AND status = 'accepted'`,
		taskID, userID,
	).Scan(&bidID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusForbidden, "no accepted bid for this task")
		return
	}

	// Verify task is assigned
	var taskStatus string
	app.db.QueryRowContext(r.Context(),
		`SELECT status FROM tasks WHERE id = ?`, taskID,
	).Scan(&taskStatus)
	if taskStatus != "assigned" {
		writeError(w, http.StatusConflict, "task is not in assigned status")
		return
	}

	// Check no pending delivery already exists
	var existingCount int
	app.db.QueryRowContext(r.Context(),
		`SELECT COUNT(*) FROM deliveries WHERE task_id = ? AND status = 'pending'`, taskID,
	).Scan(&existingCount)
	if existingCount > 0 {
		writeError(w, http.StatusConflict, "a delivery is already pending review")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		// Try as regular form
		r.ParseForm()
	}

	contentText := r.FormValue("content_text")
	var filePath, fileName string

	file, header, err := r.FormFile("file")
	if err == nil {
		defer file.Close()
		fileName = header.Filename
		ext := filepath.Ext(fileName)
		safeName := fmt.Sprintf("delivery_%d_%d_%d%s", taskID, userID, time.Now().UnixNano(), ext)
		filePath = filepath.Join(app.dataDir, "uploads", safeName)

		dst, err := os.Create(filePath)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save file")
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, file); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save file")
			return
		}
		// Store relative path
		filePath = filepath.Join("uploads", safeName)
	}

	if contentText == "" && filePath == "" {
		writeError(w, http.StatusBadRequest, "delivery must include content_text or a file")
		return
	}

	var deliveryID int64
	err = app.db.QueryRowContext(r.Context(),
		`INSERT INTO deliveries (task_id, bid_id, user_id, content_text, file_path, file_name)
		 VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
		taskID, bidID, userID, contentText, filePath, fileName,
	).Scan(&deliveryID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create delivery")
		return
	}

	// Move task to delivered
	app.db.ExecContext(r.Context(),
		`UPDATE tasks SET status = 'delivered' WHERE id = ?`, taskID)

	var d Delivery
	app.db.QueryRowContext(r.Context(),
		`SELECT id, task_id, bid_id, user_id, content_text, file_path, file_name, status, submitted_at
		 FROM deliveries WHERE id = ?`, deliveryID,
	).Scan(&d.ID, &d.TaskID, &d.BidID, &d.UserID, &d.ContentText, &d.FilePath, &d.FileName, &d.Status, &d.SubmittedAt)

	writeJSON(w, http.StatusCreated, d)
}

func (app *App) handleGetDelivery(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid task id")
		return
	}

	rows, err := app.db.QueryContext(r.Context(),
		`SELECT id, task_id, bid_id, user_id, content_text, file_path, file_name, status, submitted_at, reviewed_at
		 FROM deliveries WHERE task_id = ? ORDER BY submitted_at DESC`, taskID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()

	var deliveries []Delivery
	for rows.Next() {
		var d Delivery
		rows.Scan(&d.ID, &d.TaskID, &d.BidID, &d.UserID, &d.ContentText, &d.FilePath, &d.FileName, &d.Status, &d.SubmittedAt, &d.ReviewedAt)
		deliveries = append(deliveries, d)
	}
	if deliveries == nil {
		deliveries = []Delivery{}
	}
	writeJSON(w, http.StatusOK, deliveries)
}
