package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"laplogger/models"
)

type SwimmerHandler struct {
	db *sql.DB
}

func NewSwimmerHandler(db *sql.DB) *SwimmerHandler {
	return &SwimmerHandler{db: db}
}

func (h *SwimmerHandler) GetSwimmers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, name, email, created_at FROM swimmers ORDER BY name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var swimmers []models.Swimmer
	for rows.Next() {
		var swimmer models.Swimmer
		err := rows.Scan(&swimmer.ID, &swimmer.Name, &swimmer.Email, &swimmer.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		swimmers = append(swimmers, swimmer)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swimmers)
}

func (h *SwimmerHandler) GetSwimmer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid swimmer ID", http.StatusBadRequest)
		return
	}

	var swimmer models.Swimmer
	err = h.db.QueryRow("SELECT id, name, email, created_at FROM swimmers WHERE id = ?", id).
		Scan(&swimmer.ID, &swimmer.Name, &swimmer.Email, &swimmer.CreatedAt)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Swimmer not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(swimmer)
}

func (h *SwimmerHandler) CreateSwimmer(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSwimmerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	result, err := h.db.Exec("INSERT INTO swimmers (name, email) VALUES (?, ?)", req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created swimmer
	var swimmer models.Swimmer
	err = h.db.QueryRow("SELECT id, name, email, created_at FROM swimmers WHERE id = ?", id).
		Scan(&swimmer.ID, &swimmer.Name, &swimmer.Email, &swimmer.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(swimmer)
}
