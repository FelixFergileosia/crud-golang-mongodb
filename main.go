package main

import (
	userService "crud-golang/services"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/user", userService.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/user", userService.GetUsers).Methods(http.MethodGet)
	// router.HandleFunc("/users/{id}", userService.GetUserByID).Methods(http.MethodGet)
	// router.HandleFunc("/users/{id}", userService.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc("/user/{id}", userService.DeleteUser).Methods(http.MethodDelete)

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found—relying on real env vars")
	}
	port := os.Getenv("PORT")

	fmt.Println("Listening on port", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
