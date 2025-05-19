package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// User struct to hold user data
type User struct {
	HumanUser                   string `json:"humanUser"`
	CreateDate                  string `json:"createDate,omitempty"`
	PasswordChangedDate         string `json:"passwordChangedDate,omitempty"`
	DaysSinceLastPasswordChange int    `json:"daysSinceLastPasswordChange,omitempty"`
	LastAccessDate              string `json:"lastAccessDate,omitempty"`
	DaysSinceLastAccess         int    `json:"daysSinceLastAccess,omitempty"`
	MFAEnabled                  string `json:"mfaEnabled"`
}

// InputUser struct for POST request, omits calculated fields
type InputUser struct {
	HumanUser           string `json:"humanUser"`
	CreateDate          string `json:"createDate"`
	PasswordChangedDate string `json:"passwordChangedDate"`
	LastAccessDate      string `json:"lastAccessDate"`
	MFAEnabled          string `json:"mfaEnabled"`
}

var users []User
var db *sql.DB // Global database connection

func loadEnv() {
	err := godotenv.Load() // Load .env file from the current directory
	if err != nil {
		log.Println("No .env file found, using default or environment-set variables")
	}
}

func loadUsers() {
	connStr := os.Getenv("DB_CONN_STR")
	if connStr == "" {
		log.Fatal("DB_CONN_STR environment variable not set")
	}
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	rows, err := db.Query("SELECT human_user, create_date, password_changed_date, last_access_date, mfa_enabled FROM users_table") // Adjust table and column names
	if err != nil {
		log.Fatalf("Failed to query users from database: %v", err)
	}
	defer rows.Close()

	var loadedUsers []User
	for rows.Next() {
		var u User
		var createDate, passwordChangedDate, lastAccessDate sql.NullString
		var mfaEnabledDB sql.NullBool // Use sql.NullBool for boolean from DB

		err := rows.Scan(&u.HumanUser, &createDate, &passwordChangedDate, &lastAccessDate, &mfaEnabledDB)
		if err != nil {
			log.Printf("Failed to scan user row: %v", err)
			continue // Skip this user
		}
		// Handle nullable date strings
		if createDate.Valid {
			u.CreateDate = createDate.String
		}
		if passwordChangedDate.Valid {
			u.PasswordChangedDate = passwordChangedDate.String
		}
		if lastAccessDate.Valid {
			u.LastAccessDate = lastAccessDate.String
		}

		// Convert boolean mfaEnabledDB to "Yes"/"No" string
		if mfaEnabledDB.Valid {
			if mfaEnabledDB.Bool {
				u.MFAEnabled = "Yes"
			} else {
				u.MFAEnabled = "No"
			}
		} else {
			u.MFAEnabled = "No" // Default for NULL MFA status, or "Unknown"
		}

		loadedUsers = append(loadedUsers, u)
	}
	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating user rows: %v", err)
	}
	users = loadedUsers
	log.Printf("Successfully loaded %d users from database", len(users))
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

	if r.Method == http.MethodGet {
		getUsers(w, r)
	} else if r.Method == http.MethodPost {
		addUser(w, r)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getUsers(w http.ResponseWriter, r *http.Request) {
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

func addUser(w http.ResponseWriter, r *http.Request) {
	var newUser InputUser
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if newUser.HumanUser == "" {
		http.Error(w, "HumanUser cannot be empty", http.StatusBadRequest)
		return
	}

	var mfaEnabledDB bool
	switch strings.ToLower(newUser.MFAEnabled) {
	case "yes":
		mfaEnabledDB = true
	case "no":
		mfaEnabledDB = false
	default:
		http.Error(w, "MFAEnabled must be 'Yes' or 'No'", http.StatusBadRequest)
		return
	}

	stmt, err := db.Prepare("INSERT INTO users_table (human_user, create_date, password_changed_date, last_access_date, mfa_enabled) VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		log.Printf("Error preparing statement for add user: %v", err)
		http.Error(w, "Failed to add user", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(newUser.HumanUser, newUser.CreateDate, newUser.PasswordChangedDate, newUser.LastAccessDate, mfaEnabledDB)
	if err != nil {
		log.Printf("Error executing statement for add user: %v", err)
		http.Error(w, "Failed to add user to database", http.StatusInternalServerError)
		return
	}

	loadUsers()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User added successfully"})
	log.Printf("User %s added successfully", newUser.HumanUser)
}

func main() {
	loadEnv()
	loadUsers()

	http.HandleFunc("/api/users", usersHandler)

	port := "8080"
	log.Printf("Server starting on port %s\n", port)
	log.Printf("Users endpoint available at http://localhost:%s/api/users\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
