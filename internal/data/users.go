package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Blue-Davinci/musical-zoe/internal/database"
	"github.com/Blue-Davinci/musical-zoe/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type UserModel struct {
	DB *database.Queries
}

const (
	DefaultUserManagerDBContextTimeout = 5 * time.Second
)

var (
	ErrDuplicateEmail = errors.New("duplicate email address")
)

// Declare a new AnonymousUser variable.
var AnonymousUser = &User{}

// Check if a User instance is the AnonymousUser.
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// Create a custom password type which is a struct containing the plaintext and hashed
// versions of the password for a user.
type password struct {
	plaintext *string
	hash      []byte
}

// set() calculates the bcrypt hash of a plaintext password, and stores both
// the hash and the plaintext versions in the struct.
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// The Matches() method checks whether the provided plaintext password matches the
// hashed password stored in the struct, returning true if it matches and false
// otherwise.
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		//fmt.Printf(">>>>> Plain text: %s\nHash: %v\n", plaintextPassword, p.hash)
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

// The user struct represents a user account in our application. It contains fields for
// the user ID, created timestamp, name, email address, password hash, and activation data
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"-"`
	Version   int32     `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserSubInfo struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"created_at"`
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}
func ValidateName(v *validator.Validator, name string) {
	v.Check(name != "", "email", "must be provided")
	v.Check(len(name) <= 500, "name", "must not be more than 500 bytes long")
}
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
func ValidateUser(v *validator.Validator, user *User) {
	// Call the standalone ValidateName() helper.
	ValidateName(v, user.Name)
	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, user.Email)
	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// If the password hash is ever nil, this will be due to a logic error in our
	// codebase. So rather than adding an error to the validation map we
	// raise a panic instead.
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Insert() creates a new User and returns success on completion.
// The function will also check for the uniqueness of the user email.
// Note, this will only "Sign Up" our USER, not log them in.
func (m UserModel) Insert(user *User) error {
	// Create a new context with a 5 second timeout
	ctx, cancel := contextGenerator(context.Background(), DefaultUserManagerDBContextTimeout)
	defer cancel()
	createduser, err := m.DB.CreateUser(ctx, database.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.Password.hash,
		Activated:    user.Activated,
	})

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	// fill in the user struct with the created user data
	user.ID = createduser.ID
	user.CreatedAt = createduser.CreatedAt
	user.Version = createduser.Version
	// If we reach this point, the user was successfully created
	return nil
}

// GetForToken() retrieves a user by their API token and scope.
// It calculates the sha256 hash of the provided plaintext token, and then queries
// the database for a user with that token and scope. If found, it returns a User
// struct populated with the user's data. If not found, it returns an error.
func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// Calculate sha256 hash of plaintext
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	ctx, cancel := contextGenerator(context.Background(), DefaultUserManagerDBContextTimeout)
	defer cancel()
	// get the user
	user, err := m.DB.GetForToken(ctx, database.GetForTokenParams{
		ApiKey: tokenHash[:],
		Scope:  tokenScope,
		Expiry: time.Now(),
	})
	// check for any error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGeneralRecordNotFound
		default:
			return nil, err
		}
	}
	// make a user
	tokenuser := populateUser(user)
	// fill in the user data
	return tokenuser, nil
}

// GetByEmail() retrieves a user by their email address.
// It creates a new context with a 5 second timeout, queries the database for a user
// with the provided email, and returns a populated User struct if found. If no user
// is found, it returns an ErrGeneralRecordNotFound error. If any other error occurs,
// it returns that error.
func (m UserModel) GetByEmail(email string) (*User, error) {
	// Create a new context with a 5 second timeout
	ctx, cancel := contextGenerator(context.Background(), DefaultUserManagerDBContextTimeout)
	defer cancel()
	// get the user by email
	user, err := m.DB.GetUserByEmail(ctx, email)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrGeneralRecordNotFound
		default:
			return nil, err
		}
	}
	// populate the user struct
	populatedUser := populateUser(user)
	// no issue returning the populated user
	return populatedUser, nil
}

// UpdateUser() updates an existing user in the database.
func (m UserModel) UpdateUser(user *User) error {
	// Create a new context with a 5 second timeout
	ctx, cancel := contextGenerator(context.Background(), DefaultUserManagerDBContextTimeout)
	defer cancel()
	// Update the user in the database
	updatedUser, err := m.DB.UpdateUser(ctx, database.UpdateUserParams{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.Password.hash,
		Activated:    user.Activated,
		Version:      int32(user.Version),
	})
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	user.UpdatedAt = updatedUser.UpdatedAt
	return nil
}

// populateUser() takes a userRow of type any and attempts to convert it to a User struct.
// It checks the type of userRow, and if it is of type database.User, it creates a new
// password struct instance with the user's password hash. It then returns a pointer to a
// User struct populated with the user's ID, Name, Email, Password, Activated status,
func populateUser(userRow any) *User {
	switch user := userRow.(type) {
	case database.User:
		// Create a new password struct instance for the user.
		userPassword := password{
			hash: user.PasswordHash,
		}
		return &User{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Password:  userPassword,
			Activated: user.Activated,
			Version:   user.Version,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	default:
		// return nil if the userRow is not of type database.User
		return nil
	}
}
