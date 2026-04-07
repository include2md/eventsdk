package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type profileResponse struct {
	OK      bool        `json:"ok"`
	Profile profileData `json:"profile"`
}

type profileData struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tier      string    `json:"tier"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			id = "u-default"
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(profileResponse{
			OK: true,
			Profile: profileData{
				ID:        id,
				Name:      "Demo User",
				Tier:      "gold",
				UpdatedAt: time.Now().UTC(),
			},
		})
	})

	addr := ":18080"
	log.Printf("mock-api listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
