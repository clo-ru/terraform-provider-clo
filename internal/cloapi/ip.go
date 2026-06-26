package cloapi

import (
	"context"
	"fmt"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Address is the provider-facing view of an IP address.
type Address struct {
	ID             string
	Status         string
	Address        string
	Ptr            string
	Type           string
	CreatedIn      string
	Bandwidth      int
	IsPrimary      bool
	DdosProtection bool
	AttachedTo     *AddressAttachedTo // nil when not attached
}

// AddressAttachedTo describes the entity an address is attached to.
type AddressAttachedTo struct {
	ID     string
	Entity string
}

func addressFromSchema(r *gen.AddressSchema) Address {
	a := Address{
		ID:             r.Id,
		Status:         r.Status,
		Type:           r.Type,
		CreatedIn:      r.CreatedIn.Format(time.RFC3339),
		IsPrimary:      r.IsPrimary,
		DdosProtection: r.DdosProtection,
	}
	if r.Address != nil {
		a.Address = *r.Address
	}
	if r.Ptr != nil {
		a.Ptr = *r.Ptr
	}
	if r.BandwidthMaxMbps != nil {
		a.Bandwidth = *r.BandwidthMaxMbps
	}
	if r.AttachedTo != nil {
		a.AttachedTo = &AddressAttachedTo{ID: r.AttachedTo.Id, Entity: r.AttachedTo.Entity}
	}
	return a
}

// CreateAddress creates an address in the project and returns its ID.
func (c *Client) CreateAddress(ctx context.Context, projectID string, ddosProtection bool) (string, error) {
	body := gen.AddressCreateJSONRequestBody{}
	if ddosProtection {
		body.DdosProtection = &ddosProtection
	}
	resp, err := c.gen.AddressCreateWithResponse(ctx, projectID, body)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", fmt.Errorf("cloapi: empty address create response")
	}
	return resp.OK.Result.Id, nil
}

// GetAddress returns the address's current detail.
func (c *Client) GetAddress(ctx context.Context, id string) (*Address, error) {
	resp, err := c.gen.AddressDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, fmt.Errorf("cloapi: empty address detail response")
	}
	a := addressFromSchema(resp.OK.Result)
	return &a, nil
}

// ListAddresses returns the project's addresses (single page, matching v2 behavior).
func (c *Client) ListAddresses(ctx context.Context, projectID string) ([]Address, error) {
	resp, err := c.gen.ProjectAddressesListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Address, 0, len(items))
	for i := range items {
		out = append(out, addressFromSchema(&items[i]))
	}
	return out, nil
}

// DeleteAddress deletes the address.
func (c *Client) DeleteAddress(ctx context.Context, id string) error {
	_, err := c.gen.AddressDeleteWithResponse(ctx, id)
	return err
}

// AttachAddress attaches the address to an entity (e.g. "server" or "loadbalancer").
func (c *Client) AttachAddress(ctx context.Context, id, entityID, entityName string) error {
	_, err := c.gen.AddressAttachWithResponse(ctx, id, gen.AddressAttachJSONRequestBody{
		Id:     entityID,
		Entity: entityName,
	})
	return err
}

// DetachAddress detaches the address from its entity.
func (c *Client) DetachAddress(ctx context.Context, id string) error {
	_, err := c.gen.AddressDetachWithResponse(ctx, id)
	return err
}

// SetAddressPrimary marks the address as primary.
func (c *Client) SetAddressPrimary(ctx context.Context, id string) error {
	_, err := c.gen.AddressSetPrimaryWithResponse(ctx, id)
	return err
}

// ChangeAddressPtr sets the address PTR record.
func (c *Client) ChangeAddressPtr(ctx context.Context, id, value string) error {
	_, err := c.gen.AddressEditPtrWithResponse(ctx, id, gen.AddressEditPtrJSONRequestBody{Value: value})
	return err
}
