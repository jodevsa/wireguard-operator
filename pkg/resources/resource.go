package resources

import "context"

type Resource interface {
	Create(context.Context) error
	Update(context.Context) error
	Exists(context.Context) (bool, error)
	Converged(context.Context) (bool, error)
	NeedsUpdate(context.Context) (bool, error)
	Name() string
	Type() string
}
