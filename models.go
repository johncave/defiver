package main

import "time"

type Agent struct {
	ID                 int64     `json:"id"`
	APIKeyHash         string    `json:"-"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	CurrencyPreference string    `json:"currency_preference"`
	WalletAddress      string    `json:"wallet_address"`
	CreatedAt          time.Time `json:"created_at"`
}

type User struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"`
	DisplayName    string    `json:"display_name"`
	WalletAddress  string    `json:"wallet_address"`
	Bio            string    `json:"bio"`
	ReputationScore float64  `json:"reputation_score"`
	CreatedAt      time.Time `json:"created_at"`
}

type Task struct {
	ID          int64     `json:"id"`
	AgentID     int64     `json:"agent_id"`
	AgentName   string    `json:"agent_name,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	MaxBudget   float64   `json:"max_budget"`
	Currency    string    `json:"currency"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status"`
	BidCount    int       `json:"bid_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Bid struct {
	ID           int64     `json:"id"`
	TaskID       int64     `json:"task_id"`
	UserID       int64     `json:"user_id"`
	DisplayName  string    `json:"display_name,omitempty"`
	Amount       float64   `json:"amount"`
	DeliveryDays int       `json:"delivery_days"`
	Message      string    `json:"message"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type Delivery struct {
	ID          int64     `json:"id"`
	TaskID      int64     `json:"task_id"`
	BidID       int64     `json:"bid_id"`
	UserID      int64     `json:"user_id"`
	ContentText string    `json:"content_text"`
	FilePath    string    `json:"file_path,omitempty"`
	FileName    string    `json:"file_name,omitempty"`
	Status      string    `json:"status"`
	SubmittedAt time.Time `json:"submitted_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
}

type Payment struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	BidID     int64     `json:"bid_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
