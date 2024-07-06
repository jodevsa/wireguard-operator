package resources

import "context"

type Resource interface {
	Create(context.Context) error
	Update(context.Context) error
	Converged(ctx context.Context) (bool, error)
	NeedsUpdate(context.Context) (bool, error)
	Name() string
}
