package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/joho/godotenv"
)

// Helper function to create a mock DB and set the global db variable
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	db = mockDB // Set the global db variable to the mockDB
	return mockDB, mock
}

func TestLoadEnv(t *testing.T) {
	// Create a dummy .env file
	content := []byte("TEST_VAR=test_value")
	// Create the temp file in the current directory so Load doesn't have to guess path
	// Or, provide the full path to godotenv.Load()
	tmpDir := t.TempDir() // Create a temporary directory for the test
	tmpFilePath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(tmpFilePath, content, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Load the specific temporary .env file
	err = godotenv.Load(tmpFilePath)
	if err != nil {
		// If godotenv.Load fails when file exists, it's unexpected in this test context
		// It might print "no .env file found" if it falls back to default due to an issue.
		// For this test, we expect it to load our specific file.
		t.Logf("godotenv.Load error: %v (this might be okay if it still loads vars, but check TEST_VAR)", err)
	}

	if os.Getenv("TEST_VAR") != "test_value" {
		t.Errorf("Expected TEST_VAR to be 'test_value', got '%s'", os.Getenv("TEST_VAR"))
	}
	// Reset for other tests
	os.Unsetenv("TEST_VAR")
}

func TestLoadUsers_Success(t *testing.T) {
	mockDB, mock := newMockDB(t)
	defer mockDB.Close()

	rows := sqlmock.NewRows([]string{"human_user", "create_date", "password_changed_date", "last_access_date", "mfa_enabled"}).
		AddRow("testuser1", "Jan 1 2023", "Mar 15 2024", "May 10 2025", true).
		AddRow("testuser2", "Feb 1 2023", "Apr 20 2024", "May 15 2025", false)

	// This expectation is for the db.Query call.
	// If your main.go loadUsers has an "if db == nil" block for sql.Open and its Ping,
	// then that Ping won't be hit if 'db' (the mock) is already set.
	// If loadUsers Pings the db instance regardless, you'd add mock.ExpectPing() here.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT human_user, create_date, password_changed_date, last_access_date, mfa_enabled FROM users_table")).
		WillReturnRows(rows)

	// Use a syntactically valid (but likely non-functional) DSN.
	// This is only relevant if loadUsers calls sql.Open (e.g., if db is nil initially).
	// If your main.go's loadUsers has the "if db == nil" guard, and newMockDB sets db,
	// then sql.Open inside loadUsers might not be called.
	os.Setenv("DB_CONN_STR", "postgres://testuser:testpass@localhost:12345/testdb?sslmode=disable")
	defer os.Unsetenv("DB_CONN_STR")

	// Call loadUsers. For this test to correctly mock the Query,
	// loadUsers in main.go MUST be able to use the global 'db' (mockDB)
	// (e.g., via an 'if db == nil' guard around sql.Open).
	// If loadUsers unconditionally calls sql.Open, it overwrites the mock,
	// and this test will not correctly mock the Query.
	loadUsers()

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].HumanUser != "testuser1" || users[0].MFAEnabled != "Yes" {
		t.Errorf("Unexpected data for user1: %+v", users[0])
	}
	if users[1].HumanUser != "testuser2" || users[1].MFAEnabled != "No" {
		t.Errorf("Unexpected data for user2: %+v", users[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetUsers_Handler(t *testing.T) {
	// Setup: Ensure users slice is populated (e.g., by calling loadUsers with mock)
	mockDB, mock := newMockDB(t)
	defer mockDB.Close()
	rows := sqlmock.NewRows([]string{"human_user", "create_date", "password_changed_date", "last_access_date", "mfa_enabled"}).
		AddRow("handleruser", "Jan 1 2024", "Jan 1 2025", "May 1 2025", true)
	mock.ExpectQuery(".*").WillReturnRows(rows) // Simplified regex for any query
	os.Setenv("DB_CONN_STR", "postgres://test:test@localhost:1234/test?sslmode=disable")
	defer os.Unsetenv("DB_CONN_STR")
	loadUsers() // Populate global users slice

	req, err := http.NewRequest("GET", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	var responseUsers []User
	if err := json.Unmarshal(rr.Body.Bytes(), &responseUsers); err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	if len(responseUsers) != 1 {
		t.Errorf("Expected 1 user in response, got %d", len(responseUsers))
	}
	if responseUsers[0].HumanUser != "handleruser" {
		t.Errorf("Expected user 'handleruser', got '%s'", responseUsers[0].HumanUser)
	}
	// Add more assertions for DaysSinceLastPasswordChange, DaysSinceLastAccess if needed
}

func TestAddUser_Handler_Success(t *testing.T) {
	mockDB, mock := newMockDB(t)
	defer mockDB.Close()

	newUser := InputUser{
		HumanUser:           "newbie",
		CreateDate:          "May 18 2025",
		PasswordChangedDate: "May 18 2025",
		LastAccessDate:      "May 18 2025",
		MFAEnabled:          "Yes",
	}
	body, _ := json.Marshal(newUser)

	req, err := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Mock the INSERT statement
	mock.ExpectPrepare(regexp.QuoteMeta("INSERT INTO users_table (human_user, create_date, password_changed_date, last_access_date, mfa_enabled) VALUES ($1, $2, $3, $4, $5)")).
		ExpectExec().
		WithArgs(newUser.HumanUser, newUser.CreateDate, newUser.PasswordChangedDate, newUser.LastAccessDate, true).
		WillReturnResult(sqlmock.NewResult(1, 1)) // 1 new ID, 1 row affected

	// Mock the SELECT query from the subsequent loadUsers call
	mock.ExpectQuery(regexp.QuoteMeta("SELECT human_user, create_date, password_changed_date, last_access_date, mfa_enabled FROM users_table")).
		WillReturnRows(sqlmock.NewRows([]string{"human_user", "create_date", "password_changed_date", "last_access_date", "mfa_enabled"}).
			AddRow(newUser.HumanUser, newUser.CreateDate, newUser.PasswordChangedDate, newUser.LastAccessDate, true))

	os.Setenv("DB_CONN_STR", "postgres://test:test@localhost:1234/test?sslmode=disable") // Needed for loadUsers
	defer os.Unsetenv("DB_CONN_STR")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	expectedResponse := `{"message":"User added successfully"}`
	if strings.TrimSpace(rr.Body.String()) != expectedResponse {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expectedResponse)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddUser_Handler_InvalidInput(t *testing.T) {
	// No DB mock needed as it should fail before DB interaction
	invalidUser := InputUser{HumanUser: ""} // Empty HumanUser
	body, _ := json.Marshal(invalidUser)

	req, err := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid input: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestUsersHandler_OptionsMethod(t *testing.T) {
	req, err := http.NewRequest("OPTIONS", "/api/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("OPTIONS request returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestUsersHandler_InvalidMethod(t *testing.T) {
	req, err := http.NewRequest("PUT", "/api/users", nil) // Invalid method
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(usersHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Invalid method request returned wrong status code: got %v want %v",
			status, http.StatusMethodNotAllowed)
	}
}
