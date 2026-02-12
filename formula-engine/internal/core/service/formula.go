package service

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/huynhanx03/go-demo/formula-engine/internal/core/dto"
	"github.com/huynhanx03/go-demo/formula-engine/internal/core/entity"
	"github.com/huynhanx03/go-demo/formula-engine/internal/ports"
)

type FormulaService struct {
	formulaRepo   ports.FormulaRepository
	formulasByKey map[string]*entity.Formula
	executionPlans map[string][]string
}

func NewFormulaService(formulaRepo ports.FormulaRepository) *FormulaService {
	return &FormulaService{
		formulaRepo:    formulaRepo,
		formulasByKey:  make(map[string]*entity.Formula),
		executionPlans: make(map[string][]string),
	}
}

func (s *FormulaService) Init(ctx context.Context) error {
	formulas, err := s.formulaRepo.FindAll(ctx)
	if err != nil {
		return err
	}

	for _, f := range formulas {
		s.formulasByKey[f.Key] = f
	}

	graph, err := s.buildDependencyGraph(ctx, formulas)
	if err != nil {
		return err
	}

	globalOrder, err := s.topologicalSort(graph, s.formulasByKey)
	if err != nil {
		return err
	}

	reverseGraph := make(map[string][]string)
	for provider, consumers := range graph {
		for _, consumer := range consumers {
			reverseGraph[consumer] = append(reverseGraph[consumer], provider)
		}
	}

	for _, target := range formulas {
		needed := make(map[string]bool)

		var collectDeps func(node string)
		collectDeps = func(node string) {
			if needed[node] {
				return
			}
			needed[node] = true
			for _, provider := range reverseGraph[node] {
				collectDeps(provider)
			}
		}
		collectDeps(target.Key)

		plan := make([]string, 0, len(needed))
		for _, key := range globalOrder {
			if needed[key] {
				plan = append(plan, key)
			}
		}
		s.executionPlans[target.Key] = plan
	}

	return nil
}

func (s *FormulaService) Calculate(ctx context.Context, formulaKey string, inputVariables map[string]any) (*dto.FormulaResponse, error) {
	_, ok := s.formulasByKey[formulaKey]
	if !ok {
		return nil, fmt.Errorf("formula key not found: %s", formulaKey)
	}

	plan, ok := s.executionPlans[formulaKey]
	if !ok {
		return nil, fmt.Errorf("execution plan not found for: %s", formulaKey)
	}

	results := make(map[string]any)
	for k, v := range inputVariables {
		results[k] = v
	}

	for _, key := range plan {
		f, exists := s.formulasByKey[key]
		if !exists {
			continue
		}

		val, err := s.Evaluate(ctx, f.Expression, results)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate %s: %w", key, err)
		}
		results[key] = val
	}

	finalVal, ok := results[formulaKey]
	if !ok {
		return nil, fmt.Errorf("failed to calculate target %s", formulaKey)
	}

	return &dto.FormulaResponse{
		Value: finalVal,
		Trace: results,
	}, nil
}

func (s *FormulaService) Evaluate(ctx context.Context, expression string, variables map[string]any) (any, error) {
	program, err := expr.Compile(expression, expr.Env(variables))
	if err != nil {
		return nil, err
	}
	return expr.Run(program, variables)
}

func (s *FormulaService) buildDependencyGraph(ctx context.Context, formulas []*entity.Formula) (map[string][]string, error) {
	graph := make(map[string][]string)

	idToKey := make(map[int64]string)
	for _, f := range formulas {
		idToKey[f.ID] = f.Key
	}

	for _, f := range formulas {
		depIDs, err := s.formulaRepo.GetDependencies(ctx, f.ID)
		if err != nil {
			return nil, err
		}
		for _, depID := range depIDs {
			if childKey, ok := idToKey[depID]; ok {
				graph[childKey] = append(graph[childKey], f.Key)
			}
		}
	}
	return graph, nil
}

func (s *FormulaService) topologicalSort(graph map[string][]string, formulas map[string]*entity.Formula) ([]string, error) {
	inDegree := make(map[string]int)
	for _, f := range formulas {
		inDegree[f.Key] = 0
	}

	for _, consumers := range graph {
		for _, consumer := range consumers {
			inDegree[consumer]++
		}
	}

	queue := make([]string, 0)
	for key, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, key)
		}
	}

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, neighbor := range graph[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(formulas) {
		return nil, fmt.Errorf("cycle detected in dependencies")
	}

	return order, nil
}
