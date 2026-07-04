package pipeline

import (
	"context"
)

type Step interface {
	Name() string
	Execute(context.Context, *PipelineInput) error
}

type Processor struct {
	steps []Step
}

func NewProcessor(steps ...Step) *Processor {
	return &Processor{steps: steps}
}

func (p *Processor) Run(ctx context.Context, input *PipelineInput) error {
	for _, step := range p.steps {
		if err := step.Execute(ctx, input); err != nil {
			return err
		}
	}
	return nil
}
