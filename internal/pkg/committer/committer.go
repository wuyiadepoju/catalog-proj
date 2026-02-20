package committer

import (
	"context"

	"cloud.google.com/go/spanner"
)

// Committer defines the interface for committing transaction plans
type Committer interface {
	Apply(ctx context.Context, plan *Plan) error
}

// SpannerCommitter implements Committer using Spanner client
type SpannerCommitter struct {
	client *spanner.Client
}

// NewSpannerCommitter creates a new Spanner committer
func NewSpannerCommitter(client *spanner.Client) *SpannerCommitter {
	return &SpannerCommitter{
		client: client,
	}
}

// Apply executes all mutations in the plan atomically
func (c *SpannerCommitter) Apply(ctx context.Context, plan *Plan) error {
	if plan == nil || len(plan.Mutations()) == 0 {
		return nil
	}

	_, err := c.client.Apply(ctx, plan.Mutations())
	return err
}
