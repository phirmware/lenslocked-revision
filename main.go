package main

import (
	"fmt"
	"log"
	"net/http"
	"revision/lenslocked.com/controllers"
	"revision/lenslocked.com/models"

	"github.com/gorilla/mux"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "password"
	dbname   = "postgres"
)

func main() {
	r := mux.NewRouter()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	us, err := models.NewUserService(psqlInfo)
	staticC := controllers.NewStatic()
	usersC := controllers.NewUser(us)
	if err != nil {
		panic(err)
	}
	defer us.Close()

	us.DestructiveReset()
	// if err := us.AutoMigrate(); err != nil {
	// 	panic(err)
	// }

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("Get")
	r.HandleFunc("/signup", usersC.SignUp).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.HandleFunc("/login", usersC.SignIn).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")

	fmt.Println("Starting the server on port 3000")
	listener := http.ListenAndServe(":3000", r)
	log.Fatal(listener)
}
