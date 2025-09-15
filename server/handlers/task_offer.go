package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"mFrelance/db"
	"mFrelance/models"
)

// CreateTaskOfferHandler creates a new task offer
func CreateTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var offer models.TaskOffer
		err := json.NewDecoder(r.Body).Decode(&offer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// TODO: Add user authentication and authorization checks
		err = db.CreateTaskOffer(db.Postgres, &offer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(offer)
	}
}

// GetTaskOfferHandler retrieves a task offer by ID
func GetTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		offer, err := db.GetTaskOffer(db.Postgres, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(offer)
	}
}

// UpdateTaskOfferHandler updates an existing task offer
func UpdateTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var offer models.TaskOffer
		err = json.NewDecoder(r.Body).Decode(&offer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		offer.ID = id // Ensure the ID matches the URL
		// TODO: Add user authentication and authorization checks
		err = db.UpdateTaskOffer(db.Postgres, &offer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(offer)
	}
}

// DeleteTaskOfferHandler deletes a task offer
func DeleteTaskOfferHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// TODO: Add user authentication and authorization checks
		err = db.DeleteTaskOffer(db.Postgres, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
