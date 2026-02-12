package ports

import (
	"context"

	"github.com/huynhanx03/go-demo/formula-engine/internal/core/dto"
	"github.com/huynhanx03/go-demo/formula-engine/internal/core/entity"
)

type FormulaRepository interface {
	GetByKey(ctx context.Context, key string) (*entity.Formula, error)
	FindAll(ctx context.Context) ([]*entity.Formula, error)
	GetDependencies(ctx context.Context, formulaID int64) ([]int64, error)
}

type FormulaService interface {
	Calculate(ctx context.Context, formulaKey string, inputVariables map[string]any) (*dto.FormulaResponse, error)
	Evaluate(ctx context.Context, expression string, variables map[string]any) (any, error)
	Init(ctx context.Context) error
}
