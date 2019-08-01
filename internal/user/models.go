package user

import (
	"time"
	// "github.com/go-pg/pg/v9"
)

// User represents someone with access to our system.
type User struct {
	ID           string    `sql:"user_id,pk" json:"id"`
	Name         string    `sql:"name" json:"name"`
	Email        string    `sql:"email,unique" json:"email"`
	Roles        []string  `sql:"roles,array,notnull" json:"roles"`
	PasswordHash []byte    `sql:"password_hash" json:"-"`
	DateCreated  time.Time `sql:"date_created" json:"date_created"`
	DateUpdated  time.Time `sql:"date_updated" json:"date_updated"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string   `json:"name" validate:"required"`
	Email           string   `json:"email" validate:"required"`
	Roles           []string `json:"roles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"password_confirm" validate:"eqfield=Password"`
}

// UpdateUser defines what information may be provided to modify an existing
// User. All fields are optional so clients can send just the fields they want
// changed. It uses pointer fields so we can differentiate between a field that
// was not provided and a field that was provided as explicitly blank. Normally
// we do not want to use pointers to basic types but we make exceptions around
// marshalling/unmarshalling.
type UpdateUser struct {
	Name            *string  `json:"name"`
	Email           *string  `json:"email"`
	Roles           []string `json:"roles"`
	Password        *string  `json:"password"`
	PasswordConfirm *string  `json:"password_confirm" validate:"omitempty,eqfield=Password"`
}
