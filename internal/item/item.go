package item

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNoItemNameSupplied      = errors.New("no item name supplied")
	ErrInsufficientInformation = errors.New("insufficient information provided to create a geninue item")
)

type Item struct {
	Name     string
	ID       uuid.UUID
	Username string
	Password string
	URL      string
	Notes    []string
}

// NewItem returns an Item from the provided paramters, and additional metadata, or returns an error
func NewItem(name, username, password, url string, notes []string) (*Item, error) {
	if name == "" {
		return nil, ErrNoItemNameSupplied
	}

	if username == "" && password == "" && url == "" && notes == nil {
		return nil, ErrInsufficientInformation

	}

	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	return &Item{
		ID:       uid,
		Name:     name,
		Notes:    notes,
		Password: password,
		URL:      url,
		Username: username,
	}, nil
}
