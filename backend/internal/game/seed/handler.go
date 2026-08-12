package seed

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type seedResponse struct {
	Status string `json:"status"`
	Users  int    `json:"users"`
}

func Handler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		created, err := Seed(r.Context(), db)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"code":    "internal_error",
				"message": err.Error(),
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(seedResponse{
			Status: "seeded",
			Users:  created,
		})
	}
}
