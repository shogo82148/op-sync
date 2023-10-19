package services

import "context"

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
