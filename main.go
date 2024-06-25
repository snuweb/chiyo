package main

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/driver/mysql"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {

	fmt.Println("starting to initialize chi routes...")

	r := chi.NewRouter()
	var err error

	r.Use(middleware.Logger)
	dsn := "root:@tcp(127.0.0.1:3306)/dada?charset=utf8mb4&parseTime=True&loc=Local"

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, world"))
	})

	fmt.Println("finished initialize")
	http.ListenAndServe(":8080", r)

}
