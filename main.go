package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	verificationToken, err := generateToken()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user := User{
		Email:             email,
		Password:          string(hashedPassword),
		VerificationToken: verificationToken,
		IsVerified:        false,
	}
	// TODO: Send VerificationToken email with Mailhog
	w.Write([]byte("User registered successfull"))
}

func main() {

	fmt.Println("starting to initialize chi routes...")

	r := chi.NewRouter()
	// Database Connection
	var err error

	r.Use(middleware.Logger)
	dsn := "root:@tcp(127.0.0.1:3306)/dada?charset=utf8mb4&parseTime=True&loc=Local"

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	// migrate the scheme
	db.AutoMigrate(&User{})
	// Endpoints
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world"))
	})
	r.Post("/register", registerHandler)
	fmt.Println("finished initialize")
	http.ListenAndServe(":8080", r)

}
