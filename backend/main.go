package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"laplogger/database"
	"laplogger/handlers"
	"laplogger/middleware"
	"laplogger/models"
)

var globalDB *sql.DB

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	globalDB = db

	// Get JWT secret from environment or use default
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production.")
	}

	// Create handlers
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	swimmerHandler := handlers.NewSwimmerHandler(db)
	timeHandler := handlers.NewTimeHandler(db)

	// Create router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Public routes (no authentication required)
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")

	// Protected routes (authentication required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.JWTMiddleware(authHandler))

	// Swimmer routes
	protected.HandleFunc("/swimmers", swimmerHandler.GetSwimmers).Methods("GET")
	protected.HandleFunc("/swimmers", swimmerHandler.CreateSwimmer).Methods("POST")
	protected.HandleFunc("/swimmers/{id}", swimmerHandler.GetSwimmer).Methods("GET")

	// Time routes
	protected.HandleFunc("/times", timeHandler.CreateTime).Methods("POST")
	protected.HandleFunc("/times/{swimmer_id}", timeHandler.GetTimesBySwimmer).Methods("GET")
	protected.HandleFunc("/times", timeHandler.GetAllTimes).Methods("GET")

	// Static data routes
	protected.HandleFunc("/strokes", getStrokes).Methods("GET")
	protected.HandleFunc("/events", getEvents).Methods("GET")

	// CORS setup
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(r)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func getStrokes(w http.ResponseWriter, r *http.Request) {
	rows, err := globalDB.Query("SELECT id, name FROM strokes ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var strokes []models.Stroke
	for rows.Next() {
		var stroke models.Stroke
		err := rows.Scan(&stroke.ID, &stroke.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		strokes = append(strokes, stroke)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strokes)
}

func getEvents(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT e.id, e.stroke_id, e.distance, e.name, s.name as stroke_name
		FROM events e
		JOIN strokes s ON e.stroke_id = s.id
		ORDER BY s.id, e.distance
	`

	rows, err := globalDB.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []models.EventWithDetails
	for rows.Next() {
		var event models.EventWithDetails
		err := rows.Scan(&event.ID, &event.StrokeID, &event.Distance, &event.Name, &event.StrokeName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		events = append(events, event)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
