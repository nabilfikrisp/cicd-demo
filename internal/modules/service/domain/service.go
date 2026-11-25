package domain

import "errors"

var ErrServiceAlreadyExists = errors.New("service already exists")
var ErrInvalidService = errors.New("invalid service")

type Service struct {
	ID          string
	Name        string
	Description string
}
