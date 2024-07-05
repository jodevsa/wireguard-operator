package controllers

type Resource interface {
	Create() error
	Update() error
	Name() string
}

