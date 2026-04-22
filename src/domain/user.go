package domain

import (
	"encoding/base64"
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	id             uuid.UUID
	username       string
	hashedPassword string
	email          string
	primaryPhone   string
	createdAt      time.Time
	updatedAt      time.Time
}

func NewUser(username, password, email, primaryPhone string) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	if password == "" || len(password) < 6 {
		return nil, errors.New("password cannot be empty or less than 6 characters")
	}

	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	if primaryPhone == "" || !checkIfPhoneIsValid(primaryPhone) {
		return nil, errors.New("primary phone cannot be empty and must be valid")
	}

	if !checkIfEmailIsValid(email) {
		return nil, errors.New("email is not valid")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	return &User{
		id:             uuid.New(),
		username:       username,
		email:          email,
		primaryPhone:   primaryPhone,
		createdAt:      time.Now(),
		updatedAt:      time.Now(),
		hashedPassword: hashedPassword,
	}, nil
}

// ReconstructUser creates a User from persisted data, bypassing validation and password hashing.
func ReconstructUser(id uuid.UUID, username, hashedPassword, email, primaryPhone string, createdAt, updatedAt time.Time) User {
	return User{
		id:             id,
		username:       username,
		hashedPassword: hashedPassword,
		email:          email,
		primaryPhone:   primaryPhone,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (u User) GetID() uuid.UUID {
	return u.id
}
func (u User) GetUsername() string {
	return u.username
}
func (u User) GetEmail() string {
	return u.email
}
func (u User) GetPrimaryPhone() string {
	return u.primaryPhone
}
func (u User) GetCreatedAt() time.Time {
	return u.createdAt
}
func (u User) GetUpdatedAt() time.Time {
	return u.updatedAt
}
func (u User) GetHashedPassword() string {
	return u.hashedPassword
}

func (u *User) SetID(id uuid.UUID) {
	u.id = id
}
func (u *User) SetEmail(email string) error {
	if email == "" || !checkIfEmailIsValid(email) {
		return errors.New("email cannot be empty and must be valid")
	}
	u.email = email
	return nil
}
func (u *User) SetPrimaryPhone(phone string) error {
	if phone == "" || !checkIfPhoneIsValid(phone) {
		return errors.New("primary phone cannot be empty and must be valid")
	}
	u.primaryPhone = phone
	return nil
}
func (u *User) SetHashedPassword(hashedPassword string) {
	u.hashedPassword = hashedPassword
}
func (u *User) SetUpdatedAt(date time.Time) {
	u.updatedAt = date
}
func (u *User) SetCreatedAt(date time.Time) {
	u.createdAt = date
}

func (u User) CheckPassword(password string) bool {
	decryptedPassword, err := base64.StdEncoding.DecodeString(u.hashedPassword)
	if err != nil {
		return false
	}
	err = bcrypt.CompareHashAndPassword(decryptedPassword, []byte(password))
	return err == nil
}

func checkIfEmailIsValid(email string) bool {
	// Simple regex for email validation
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
func checkIfPhoneIsValid(phone string) bool {
	// Simple regex for phone number validation (example: US phone numbers)
	re := regexp.MustCompile(`^\+?1?\d{10,15}$`)
	return re.MatchString(phone)
}
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hashPassword := base64.StdEncoding.EncodeToString(bytes)

	return hashPassword, nil
}
