package controllers

import (
	"fmt"
	"net/http"
	"revision/lenslocked.com/models"
	"revision/lenslocked.com/views"
)

type User struct {
	SignUpView *views.View
	us         *models.UserService
}

type SignUpForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

func NewUser(us *models.UserService) *User {
	return &User{
		SignUpView: views.NewView("bootstrap", "users/signup"),
		us: us,
	}
}

func (u *User) SignUp(w http.ResponseWriter, r *http.Request) {
	u.SignUpView.Render(w, nil)
}

func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	var form SignUpForm

	if err := ParseForm(r, &form); err != nil {
		panic(err)
	}

	user := models.User{
		Name:  form.Name,
		Email: form.Email,
		Password: form.Password,
	}

	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	foundUser, err := u.us.ByEmail(user.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "A user with Name %s , Email %s, and password hash %s", foundUser.Name, foundUser.Email, foundUser.PasswordHash)

}
