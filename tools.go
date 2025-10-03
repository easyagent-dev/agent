package agent

import "context"

type ModelTool interface {
	Name() string

	Description() string

	InputSchema() any

	OutputSchema() any

	Run(ctx context.Context, input any) (any, error)

	Usage() string
}
