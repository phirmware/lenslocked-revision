package controllers

import (
	"fmt"
	"net/http"
	"revision/lenslocked.com/models"
	"revision/lenslocked.com/rand"
	"revision/lenslocked.com/views"
)

// User defines the shape of the user controller
type User struct {
	SignUpView *views.View
	LoginView  *views.View
	us         models.UserService
}

// SignUpForm defines the shape of the signup form
type SignUpForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// LoginForm defines the shape of the login form
type LoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// NewUser returns the usermodel struct
func NewUser(us models.UserService) *User {
	return &User{
		SignUpView: views.NewView("bootstrap", "users/signup"),
		LoginView:  views.NewView("bootstrap", "users/login"),
		us:         us,
	}
}

// SignUp handles the GET signup
func (u *User) SignUp(w http.ResponseWriter, r *http.Request) {
	u.SignUpView.Render(w, nil)
}

// Create handles the POST signup
func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	var form SignUpForm

	if err := ParseForm(r, &form); err != nil {
		panic(err)
	}

	user := models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}

	if err := u.us.Create(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// foundUser, err := u.us.ByEmail(user.Email)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	if err := u.setCookie(w, &user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cookie", http.StatusFound)
}

// SignIn handles the GET login
func (u *User) SignIn(w http.ResponseWriter, r *http.Request) {
	u.LoginView.Render(w, nil)
}

// Login handles the POST login
func (u *User) Login(w http.ResponseWriter, r *http.Request) {
	var user LoginForm
	if err := ParseForm(r, &user); err != nil {
		panic(err)
	}

	foundUser, err := u.us.Authenticate(user.Email, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := u.setCookie(w, foundUser); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/cookie", http.StatusFound)
}

// CookieTest gets the cookie and returns it
func (u *User) CookieTest(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("remember_token")
	if err != nil {
		panic(err)
	}

	fmt.Println(cookie.Value, "Cookit value")

	user, err := u.us.ByRemember(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%+v", user)
}

func (u *User) setCookie(w http.ResponseWriter, user *models.User) error {
	if user.Remember == "" {
		str, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = str
		u.us.Update(user)
	}
	fmt.Println(user.Remember, "user remember token")
	cookie := http.Cookie{
		Name:     "remember_token",
		Value:    user.Remember,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	return nil
}
