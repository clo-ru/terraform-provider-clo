package cloapi

import (
	"context"
	"time"
)

// Project is the provider-facing view of a project.
type Project struct {
	ID             string
	Name           string
	Status         string
	CreatedIn      string
	StoppingReason string
	HasAbuse       bool
}

// ListProjects returns the account's projects (single page, matching v2 behavior).
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	resp, err := c.gen.ProjectListWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Project, 0, len(items))
	for _, p := range items {
		proj := Project{
			ID:        p.Id,
			Name:      p.Name,
			Status:    p.Status,
			CreatedIn: p.Created.Format(time.RFC3339),
		}
		if p.HasAbuse != nil {
			proj.HasAbuse = *p.HasAbuse
		}
		if p.StoppingReason != nil {
			proj.StoppingReason = *p.StoppingReason
		}
		out = append(out, proj)
	}
	return out, nil
}
