package entity

import (
	"time"
)

type Formula struct {
	ID          int64
	Key         string
	Expression  string
	Description string
	IsPublic    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewFormula(id int64, key, expression, description string, isPublic bool) (*Formula, error) {
	return &Formula{
		ID:          id,
		Key:         key,
		Expression:  expression,
		Description: description,
		IsPublic:    isPublic,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
