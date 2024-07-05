package resources

import "context"

type Resource interface {
	Create(context.Context) error
	Update( context.Context) error
	Name() string
}