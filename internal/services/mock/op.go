package mock

import (
	"context"

	"github.com/shogo82148/op-sync/internal/services"
)

var _ services.WhoAmIer = WhoAmIer(nil)

// WhoAmIer returns the information about a signed-in account.
type WhoAmIer func(ctx context.Context) (*services.OnePasswordUser, error)

func (f WhoAmIer) WhoAmI(ctx context.Context) (*services.OnePasswordUser, error) {
	return f(ctx)
}

var _ services.Injector = Injector(nil)

// Injector inject the secrets into the template.
type Injector func(ctx context.Context, template string) ([]byte, error)

func (f Injector) Inject(ctx context.Context, template string) ([]byte, error) {
	return f(ctx, template)
}

var _ services.OnePasswordItemGetter = OnePasswordItemGetter(nil)

// OnePasswordItemGetter gets the item from 1password.
type OnePasswordItemGetter func(ctx context.Context, repository, name string) (*services.OnePasswordItem, error)

func (f OnePasswordItemGetter) GetOnePasswordItem(ctx context.Context, repository, name string) (*services.OnePasswordItem, error) {
	return f(ctx, repository, name)
}

var _ services.OnePasswordReader = OnePasswordReader(nil)

// OnePasswordReader reads the secret from 1password.
type OnePasswordReader func(ctx context.Context, uri string) ([]byte, error)

func (f OnePasswordReader) ReadOnePassword(ctx context.Context, uri string) ([]byte, error) {
	return f(ctx, uri)
}
