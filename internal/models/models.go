package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/romeulima/devbook-server/internal/security"
)

type Response struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type UserPayload struct {
	Name     string `json:"name,omitempty"`
	Nick     string `json:"nick,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Nick      string    `json:"nick"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewUserResponse(id uuid.UUID, up UserPayload, createdAt time.Time) *UserResponse {
	return &UserResponse{
		ID:        id,
		Name:      up.Name,
		Nick:      up.Nick,
		Email:     up.Email,
		CreatedAt: createdAt,
	}
}

type MissingFieldError struct {
	Field   string
	Message string
}

func (up *UserPayload) PrepareUser(payload *UserPayload) error {
	err := up.ValidateFields(*payload, "cadastro")

	if err != nil {
		return err
	}

	encondedPassword, err := security.EncryptPassword(payload.Password)

	if err != nil {
		return err
	}

	payload.Password = string(encondedPassword)

	return nil
}

func (e *MissingFieldError) Error() string {
	return fmt.Sprintf("the field %s %s", e.Field, e.Message)
}

func (ps *UserPayload) ValidateFields(payload UserPayload, mode string) error {
	switch {
	case payload.Name == "":
		return &MissingFieldError{Field: "name", Message: "is required"}
	case payload.Nick == "":
		return &MissingFieldError{Field: "nick", Message: "is required"}
	case payload.Email == "":
		return &MissingFieldError{Field: "email", Message: "is required"}
	case payload.Password == "" && mode == "cadastro":
		return &MissingFieldError{Field: "password", Message: "is required"}
	}
	return nil
}
