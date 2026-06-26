package cloapi

import "context"

// Image is the provider-facing view of an OS image.
type Image struct {
	ID             string
	Name           string
	OSDistribution string
	OSFamily       string
	OSVersion      string
}

// ListImages returns the project's OS images (single page, matching v2 behavior).
func (c *Client) ListImages(ctx context.Context, projectID string) ([]Image, error) {
	resp, err := c.gen.ProjectImagesListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Image, 0, len(items))
	for _, im := range items {
		out = append(out, Image{
			ID:             im.Id,
			Name:           im.Name,
			OSDistribution: im.OperationSystem.Distribution,
			OSFamily:       im.OperationSystem.OsFamily,
			OSVersion:      im.OperationSystem.Version,
		})
	}
	return out, nil
}
