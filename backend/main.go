package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
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
