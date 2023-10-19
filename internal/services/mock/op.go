package mock

import (
	"context"

	"github.com/shogo82148/op-sync/internal/services"
)

var _ services.WhoAmIer = WhoAmIer(nil)

type WhoAmIer func(ctx context.Context) (*services.OnePasswordUser, error)

func (f WhoAmIer) WhoAmI(ctx context.Context) (*services.OnePasswordUser, error) {
	return f(ctx)
}

var _ services.Injector = Injector(nil)

type Injector func(ctx context.Context, template string) ([]byte, error)

func (f Injector) Inject(ctx context.Context, template string) ([]byte, error) {
	return f(ctx, template)
}
