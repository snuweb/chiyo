package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func newmailesender(email string, token string) error {

}

func generateNewToken() (string, error) {
	// make slice of 16 length
	bytes := make([]byte, 16)

	// fill the slice and check the error at the same time
	if _, err := rand.Read(bytes); err != nil {

		return "", err

	}

	return hex.EncodeToString(bytes), nil
}

func sendVerificationEmail(email string, token string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@example.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Please verify your email address")
	m.SetBody("text/plain", fmt.Sprintf("Please click the following link to verify your email address: http://localhost:8080/verify?token=%s", token))

	// MailHog SMTP server should use port 1025
	d := gomail.NewDialer("localhost", 1025, "", "")

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// send email

// func sendVerificationEmailTwo(email string, token string) error {
// 	m := gomail.NewMessage()
// 	m.SetHeader("From", "no-reply@myganacsi.com")
// 	m.SetHeader("To", email)
// 	m.SetHeader("Subject", "Please verify your email address")
// 	m.SetBody("text/plain", fmt.Sprintf("Please click the flowing link to verify your email address: localhost:8080/verify?token=%s", token))

// 	// mailHog SMTO server should use port 1025
// 	d := gomail.NewDialer("localhost", 1025, "", "")

// 	if err := d.DialAndSend(m); err != nil {
// 		return err
// 	}
// 	return nil
// }

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		log.Printf("Err", err)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	user.IsVerified = true
	user.VerificationToken = ""

	if err := db.Save(&user).Error; err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, "could not verify user", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Email verified successfully"))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	verificationToken, err := generateToken()
	if err != nil {
		log.Printf("Error generating token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := User{
		Email:             email,
		Password:          string(hashedPassword),
		VerificationToken: verificationToken,
		IsVerified:        false,
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	if err := sendVerificationEmail(email, verificationToken); err != nil {
		log.Printf("Error sending verification email: %v", err)
		http.Error(w, "Could not send verification email", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User registered successfully"))
}
func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/login.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var user User
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("Error finding user: %v", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if !user.IsVerified {
		http.Error(w, "Email not verified", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Printf("Error comparing password: %v", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	w.Write([]byte("Login successful"))
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Database connection
	var err error
	dsn := "root:@tcp(127.0.0.1:3306)/dada?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}

	log.Println("Connected to the database")

	// Migrate the schema
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("Could not migrate the database: %v", err)
	}

	log.Println("Database migrated successfully")

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world!"))
	})

	r.Post("/register", registerHandler)
	r.Get("/verify", verifyHandler)

	r.Post("/register", registerHandler)
	r.Get("/verify", verifyHandler)
	r.Get("/login", loginFormHandler)
	r.Post("/login", loginHandler)

	log.Println("Starting server on :8080")
	http.ListenAndServe(":8080", r)
}
