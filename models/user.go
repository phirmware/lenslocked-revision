package models

import (
	"errors"
	"fmt"

	"revision/lenslocked.com/hash"

	"revision/lenslocked.com/rand"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
)

const (
	hmacSecretKey = "hmac-secret-key"
)

var (
	// ErrNotFound is returned no resource exists
	ErrNotFound = errors.New("models: resource not found")

	// ErrInvalidID is returned when an invalid ID e.g 0 is provided
	ErrInvalidID = errors.New("models: invalid id provided")

	// ErrPasswordMissing ris returned when the password is empty
	ErrPasswordMissing = errors.New("models: Password is missing")

	// ErrUserNotFound is returned when the user is not in the db
	ErrUserNotFound = errors.New("models: User not found")

	// ErrInvalidPassword is returned when the password entered for a particular user is wrong
	ErrInvalidPassword = errors.New("models: Invalid password")
)

// User represents a user object
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// UserDB is the interface for the database layer
type UserDB interface {
	ByID(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)
	Delete(id uint) error
	Find() (*[]User, error)
	Create(user *User) error
	Update(user *User) error
	AutoMigrate() error
	Close() error
	DestructiveReset() error
}

// UserService is the interface tyor for the models
type UserService interface {
	UserDB
	Authenticate(email, password string) (*User, error)
}

// UserService defines the service
type userService struct {
	UserDB
}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

type userValidator struct {
	UserDB
	hmac hash.HMAC
}

type userValFns func(user *User) error

var _ UserDB = &userGorm{}
var _ UserDB = &userValidator{}
var _ UserService = &userService{}

func runValFns(user *User, fns ...userValFns) error {
	for _, fn := range(fns) {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func newUserGorm(connectionInfo string) (UserDB, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &userGorm{
		db: db,
	}, nil
}

func newUserValidators(connectionInfo string) (*userValidator, error) {
	ug, err := newUserGorm(connectionInfo)
	hmac := hash.NewHmac(hmacSecretKey)
	if err != nil {
		return nil, err
	}
	return &userValidator{
		UserDB: ug,
		hmac:   hmac,
	}, nil
}

// NewUserService returns a UserService
func NewUserService(connectionInfo string) (UserService, error) {
	uv, err := newUserValidators(connectionInfo)
	if err != nil {
		return nil, err
	}
	return &userService{
		UserDB: uv,
	}, nil
}

// ByID looks up a user by Id
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail finds a user by email
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ByRemember validator function
func (uv *userValidator) ByRemember(rememberToken string) (*User, error) {
	hashedToken := uv.hmac.Hash(rememberToken)
	return uv.UserDB.ByRemember(hashedToken)
}

// ByRemember searches a user by remember hash
func (ug *userGorm) ByRemember(hashedToken string) (*User, error) {
	var user User
	if err := first(ug.db.Where("remember_hash = ?", hashedToken), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete for user validation
func (uv *userValidator) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	return uv.UserDB.Delete(id)
}

// Delete removes a user from the database
func (ug *userGorm) Delete(id uint) error {
	var user User
	user.ID = id
	return ug.db.Delete(&user).Error
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

// Find gets all users
func (ug *userGorm) Find() (*[]User, error) {
	var users []User
	if err := ug.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return &users, nil
}

func (uv *userValidator) bycryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}
	hashedByte, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedByte)
	user.Password = ""
	return nil
}

// Create validator function
func (uv *userValidator) Create(user *User) error {
	if err := runValFns(user, uv.bycryptPassword); err != nil {
		return err
	}

	if user.Remember == "" {
		remember, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = remember
	}
	return uv.UserDB.Create(user)
}

// Create a new user
func (ug *userGorm) Create(user *User) error {
	hashedByte, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedByte)
	user.Password = ""
	if user.Remember == "" {
		user.Remember, err = rand.RememberToken()
		if err != nil {
			return err
		}
	}
	user.RememberHash = ug.hmac.Hash(user.Remember)
	fmt.Printf("%+v", user)
	return ug.db.Create(&user).Error
}

// Authenticate authenticates the user on login
func (us *userService) Authenticate(email, password string) (*User, error) {
	if password == "" {
		return nil, ErrPasswordMissing
	}

	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidPassword
	}

	return foundUser, nil
}

// Update is uservalidator f
func (uv *userValidator) update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = uv.hmac.Hash(user.Remember)
	}
	return uv.UserDB.Update(user)
}

// Update will update the user
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Close closes connections to the database
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

// DestructiveReset drops and creates a table
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// AutoMigrate migrates the user model to the db
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}
