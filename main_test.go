package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// TestGenerateToken tests the generateToken function
func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(token) != 32 {
		t.Fatalf("Expected token length of 32, got %d", len(token))
	}
}

// TestSendVerificationEmail tests the sendVerificationEmail function
func TestSendVerificationEmail(t *testing.T) {
	token := "testtoken"
	email := "test@example.com"
	err := sendVerificationEmail(email, token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

// Helper function to initialize the database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "root:@tcp(127.0.0.1:3306)/dada_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Could not connect to the database: %v", err)
	}
	err = db.AutoMigrate(&User{})
	if err != nil {
		t.Fatalf("Could not migrate the database: %v", err)
	}
	return db
}

// TestRegisterHandler tests the registerHandler function
func TestRegisterHandler(t *testing.T) {
	db = setupTestDB(t)

	data := url.Values{}
	data.Set("email", "test@example.com")
	data.Set("password", "password")

	req, err := http.NewRequest("POST", "/register", strings.NewReader(data.Encode()))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(registerHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, status)
	}

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("Could not read response body: %v", err)
	}

	expected := "User registered successfully"
	if strings.TrimSpace(string(body)) != expected {
		t.Fatalf("Expected response body %q, got %q", expected, body)
	}
}

// TestLoginHandler tests the loginHandler function
func TestLoginHandler(t *testing.T) {
	db = setupTestDB(t)

	// Create a test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Could not hash password: %v", err)
	}

	user := User{
		Email:             "test@example.com",
		Password:          string(hashedPassword),
		VerificationToken: "",
		IsVerified:        true,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Could not create user: %v", err)
	}

	data := url.Values{}
	data.Set("email", "test@example.com")
	data.Set("password", "password")

	req, err := http.NewRequest("POST", "/login", strings.NewReader(data.Encode()))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(loginHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, status)
	}

	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("Could not read response body: %v", err)
	}

	expected := "Login successful"
	if strings.TrimSpace(string(body)) != expected {
		t.Fatalf("Expected response body %q, got %q", expected, body)
	}
}
