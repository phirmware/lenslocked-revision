package main

import (
	"fmt"
	"revision/lenslocked.com/models"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	const (
		host     = "localhost"
		port     = 5432
		user     = "postgres"
		password = "password"
		dbname   = "postgres"
	)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	us, err := models.NewUserService(psqlInfo)
	if err != nil {
		panic(err)
	}

	// us.DestructiveReset()

	// testStuff(us)
	// findEmail(us)
	// byID(us)
	deleteUser(us)
	
}

func testStuff(us *models.UserService) {
	user := models.User{
		Name: "phirmware",
		Email: "Phirmware@mail.com",
	}
	if err := us.Create(&user); err != nil {
		panic(err)
	}
	fmt.Printf("%+v", user)
}

func findEmail(us *models.UserService) {
	email := "Phirmware@mail.com"
	user, err := us.ByEmail(email)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", user)
}

func byID(us *models.UserService) {
	id := 1
	user, err := us.ByID(uint(id))
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", user)
}

func deleteUser(us *models.UserService) {
	id := 1
	if err := us.Delete(uint(id)); err != nil {
		panic(err)
	}
	
	byID(us)
}

