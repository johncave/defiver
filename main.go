package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed frontend/dist
var frontendFS embed.FS

type App struct {
	db        *sql.DB
	jwtSecret string
	dataDir   string
}

func main() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := openDB(dataDir)
	defer db.Close()

	app := &App{db: db, jwtSecret: jwtSecret, dataDir: dataDir}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Agent API
	r.Post("/api/v1/agents/register", app.handleAgentRegister)
	r.Group(func(r chi.Router) {
		r.Use(app.requireAgent)
		r.Get("/api/v1/agents/me", app.handleAgentMe)
		r.Post("/api/v1/agents/tasks", app.handleAgentCreateTask)
		r.Get("/api/v1/agents/tasks", app.handleAgentListTasks)
		r.Get("/api/v1/agents/tasks/{id}", app.handleAgentGetTask)
		r.Post("/api/v1/agents/tasks/{id}/bids/{bid_id}/accept", app.handleAgentAcceptBid)
		r.Post("/api/v1/agents/tasks/{id}/delivery/review", app.handleAgentReviewDelivery)
	})

	// User auth
	r.Post("/api/v1/users/register", app.handleUserRegister)
	r.Post("/api/v1/users/login", app.handleUserLogin)

	// Public task endpoints
	r.Get("/api/v1/tasks", app.handlePublicListTasks)
	r.Get("/api/v1/tasks/{id}", app.handlePublicGetTask)
	r.Get("/api/v1/tasks/{id}/bids", app.handlePublicListBids)

	// Authenticated user endpoints
	r.Group(func(r chi.Router) {
		r.Use(app.requireUser)
		r.Get("/api/v1/users/me", app.handleUserMe)
		r.Post("/api/v1/tasks/{id}/bids", app.handleSubmitBid)
		r.Get("/api/v1/users/bids", app.handleMyBids)
		r.Delete("/api/v1/users/bids/{bid_id}", app.handleWithdrawBid)
		r.Post("/api/v1/tasks/{id}/deliveries", app.handleSubmitDelivery)
		r.Get("/api/v1/tasks/{id}/deliveries", app.handleGetDelivery)
	})

	// Serve uploaded files
	r.Get("/uploads/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(dataDir+"/uploads"))).ServeHTTP(w, r)
	})

	// Serve Vue SPA — all non-API routes
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("failed to sub frontend/dist: %v", err)
	}
	fileServer := http.FileServer(http.FS(distFS))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file; fall back to index.html for SPA routing
		_, err := distFS.Open(r.URL.Path[1:])
		if err != nil {
			http.ServeFileFS(w, r, distFS, "index.html")
			return
		}
		fileServer.ServeHTTP(w, r)
	})

	log.Printf("Defiver listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
