package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// User struct to hold user data
type User struct {
	HumanUser                   string `json:"humanUser"`
	CreateDate                  string `json:"createDate"`
	PasswordChangedDate         string `json:"passwordChangedDate"`
	DaysSinceLastPasswordChange int    `json:"daysSinceLastPasswordChange"`
	LastAccessDate              string `json:"lastAccessDate"`
	DaysSinceLastAccess         int    `json:"daysSinceLastAccess"`
	MFAEnabled                  string `json:"mfaEnabled"`
}

var users []User

func loadUsers() {
	data, err := os.ReadFile("users.json")
	if err != nil {
		log.Fatalf("Failed to read users.json: %v", err)
	}
	err = json.Unmarshal(data, &users)
	if err != nil {
		log.Fatalf("Failed to unmarshal users data: %v", err)
	}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	responseUsers := make([]User, len(users))
	copy(responseUsers, users) // Work on a copy to ensure calculations are fresh per request

	now := time.Now().UTC()
	dateFormat := "Jan 2 2006"
	for i := range responseUsers {
		currentUser := &responseUsers[i]
		// Calculate DaysSinceLastPasswordChange
		if currentUser.PasswordChangedDate != "" {
			pwdChangedDate, err := time.Parse(dateFormat, currentUser.PasswordChangedDate)
			if err != nil {
				log.Printf("Error parsing PasswordChangedDate ('%s') for user %s: %v", currentUser.PasswordChangedDate, currentUser.HumanUser, err)
				currentUser.DaysSinceLastPasswordChange = -1 // Indicate error
			} else {
				duration := now.Sub(pwdChangedDate)
				currentUser.DaysSinceLastPasswordChange = int(duration.Hours() / 24)
			}
		} else {
			currentUser.DaysSinceLastPasswordChange = -1 // Indicate missing date
		}

		// Calculate DaysSinceLastAccess
		if currentUser.LastAccessDate != "" {
			lastAccess, err := time.Parse(dateFormat, currentUser.LastAccessDate)
			if err != nil {
				log.Printf("Error parsing LastAccessDate ('%s') for user %s: %v", currentUser.LastAccessDate, currentUser.HumanUser, err)
				currentUser.DaysSinceLastAccess = -1 // Indicate error
			} else {
				duration := now.Sub(lastAccess)
				currentUser.DaysSinceLastAccess = int(duration.Hours() / 24)
			}
		} else {
			currentUser.DaysSinceLastAccess = -1 // Indicate missing date
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responseUsers)
}

func main() {
	loadUsers()

	http.HandleFunc("/api/users", usersHandler)

	port := "8080"
	log.Printf("Server starting on port %s\n", port)
	log.Printf("Users endpoint available at http://localhost:%s/api/users\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
