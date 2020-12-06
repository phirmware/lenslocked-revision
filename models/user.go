package models

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
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

	// ErrInvalidPassword is returned when the password entered for a oparticular user is wrong
	ErrInvalidPassword = errors.New("models: Invalid password")
)

// User represents a user object
type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
}

// UserService defines the service
type UserService struct {
	db *gorm.DB
}

// NewUserService returns a UserService
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &UserService{
		db: db,
	}, nil
}

// ByID looks up a user by Id
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// ByEmail finds a user by email
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	if err := first(db, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// Delete removes a user from the database
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	var user User
	user.ID = id
	return us.db.Delete(&user).Error
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}

// Find gets all users
func (us *UserService) Find() (*[]User, error) {
	var users []User
	if err := us.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return &users, nil
}

// Create a new user
func (us *UserService) Create(user *User) error {
	hashedByte, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedByte)
	user.Password = ""
	return us.db.Create(&user).Error
}

// Authenticate authenticates the user on login
func (us *UserService) Authenticate(email, password string) (*User, error) {
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

// Update will update the user
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

// Close closes connections to the database
func (us *UserService) Close() error {
	return us.db.Close()
}

// DestructiveReset drops and creates a table
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

// AutoMigrate migrates the user model to the db
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}
