package services

import (
	"context"
	"time"
)

type OnePasswordUser struct {
	URL         string `json:"url"`
	Email       string `json:"email"`
	UserUUID    string `json:"user_uuid"`
	AccountUUID string `json:"account_uuid"`
	Shorthand   string `json:"shorthand"`
}

// WhoAmIer returns the information about a signed-in account.
type WhoAmIer interface {
	WhoAmI(ctx context.Context) (*OnePasswordUser, error)
}

// Injector inject the secrets into the template.
type Injector interface {
	Inject(ctx context.Context, template string) ([]byte, error)
}

type OnePasswordItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Version int    `json:"version"`
	Vault   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"vault"`
	Category              string    `json:"category"`
	LastEditedBy          string    `json:"last_edited_by"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	AdditionalInformation string    `json:"additional_information"`
}

// OnePasswordItemGetter gets the item from 1password.
type OnePasswordItemGetter interface {
	GetOnePasswordItem(ctx context.Context, vault, item string) (*OnePasswordItem, error)
}

// OnePasswordReader reads the secret from 1password.
type OnePasswordReader interface {
	ReadOnePassword(ctx context.Context, uri string) ([]byte, error)
}
