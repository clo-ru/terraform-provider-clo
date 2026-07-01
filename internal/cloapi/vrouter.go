package cloapi

import (
	"context"
	"errors"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Vrouter is the provider-facing view of a virtual router. Status is the
// provisioning state; SwitchStatus is the power state toggled by Start/Stop.
type Vrouter struct {
	ID                       string
	Name                     string
	Project                  string
	Status                   string
	SwitchStatus             string
	ExternalGatewayAddressID string
	PrivateNetworks          []string
}

func vrouterFromSchema(r *gen.VrouterSchema) Vrouter {
	v := Vrouter{
		ID:           r.Id,
		Name:         r.Name,
		Project:      r.Project,
		Status:       r.Status,
		SwitchStatus: r.SwitchStatus,
	}
	if r.ExternalGatewayAddressId != nil {
		v.ExternalGatewayAddressID = *r.ExternalGatewayAddressId
	}
	if r.PrivateNetworks != nil {
		v.PrivateNetworks = append([]string(nil), *r.PrivateNetworks...)
	}
	return v
}

// VrouterCreateParams holds the inputs for creating a virtual router.
type VrouterCreateParams struct {
	Name            string
	PrivateNetworks []string
}

// CreateVrouter creates a virtual router in the project and returns its ID.
func (c *Client) CreateVrouter(ctx context.Context, projectID string, p VrouterCreateParams) (string, error) {
	body := gen.VrouterCreateJSONRequestBody{Name: p.Name}
	if len(p.PrivateNetworks) > 0 {
		nets := append([]string(nil), p.PrivateNetworks...)
		body.PrivateNetworks = &nets
	}
	resp, err := c.gen.VrouterCreateWithResponse(ctx, projectID, body)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty vrouter create response")
	}
	return resp.OK.Result.Id, nil
}

// GetVrouter returns the virtual router's current detail.
func (c *Client) GetVrouter(ctx context.Context, id string) (*Vrouter, error) {
	resp, err := c.gen.VrouterDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty vrouter detail response")
	}
	v := vrouterFromSchema(resp.OK.Result)
	return &v, nil
}

// ListVrouters returns the project's virtual routers (single page, matching the other list adapters).
func (c *Client) ListVrouters(ctx context.Context, projectID string) ([]Vrouter, error) {
	resp, err := c.gen.ProjectVrouterListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Vrouter, 0, len(items))
	for i := range items {
		out = append(out, vrouterFromSchema(&items[i]))
	}
	return out, nil
}

// StartVrouter powers the virtual router on.
func (c *Client) StartVrouter(ctx context.Context, id string) error {
	_, err := c.gen.VrouterStartWithResponse(ctx, id)
	return err
}

// StopVrouter powers the virtual router off.
func (c *Client) StopVrouter(ctx context.Context, id string) error {
	_, err := c.gen.VrouterStopWithResponse(ctx, id)
	return err
}

// DeleteVrouter deletes a virtual router.
func (c *Client) DeleteVrouter(ctx context.Context, id string) error {
	_, err := c.gen.VrouterDeleteWithResponse(ctx, id)
	return err
}
