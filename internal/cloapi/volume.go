package cloapi

import (
	"context"
	"fmt"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Volume is the provider-facing view of a volume. Adapter types deliberately do not
// expose generated SDK types, so generated name/shape changes stay inside this package.
type Volume struct {
	ID         string
	Status     string
	CreatedIn  string
	Attachment *VolumeAttachment // nil when the volume is not attached
}

// VolumeAttachment describes where a volume is attached.
type VolumeAttachment struct {
	Device string
	ID     string
}

// VolumeCreateParams describes a new volume.
type VolumeCreateParams struct {
	ProjectID  string
	Name       string
	Size       int
	Autorename bool
}

// CreateVolume creates a volume and returns its ID.
func (c *Client) CreateVolume(ctx context.Context, p VolumeCreateParams) (string, error) {
	body := gen.VolumeCreateJSONRequestBody{Name: p.Name, Size: p.Size}
	if p.Autorename {
		auto := true
		body.Autorename = &auto
	}
	resp, err := c.gen.VolumeCreateWithResponse(ctx, p.ProjectID, body)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", fmt.Errorf("cloapi: empty volume create response")
	}
	return resp.OK.Result.Id, nil
}

// GetVolume returns the volume's current detail.
func (c *Client) GetVolume(ctx context.Context, id string) (*Volume, error) {
	resp, err := c.gen.VolumeDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, fmt.Errorf("cloapi: empty volume detail response")
	}
	r := resp.OK.Result
	v := &Volume{
		ID:        r.Id,
		Status:    r.Status,
		CreatedIn: r.CreatedIn.Format(time.RFC3339),
	}
	if r.AttachedToServer != nil {
		v.Attachment = &VolumeAttachment{Device: r.AttachedToServer.Device, ID: r.AttachedToServer.Id}
	}
	return v, nil
}

// AttachVolume attaches the volume to a server.
func (c *Client) AttachVolume(ctx context.Context, volumeID, serverID string) error {
	_, err := c.gen.VolumeAttachWithResponse(ctx, volumeID, gen.VolumeAttachJSONRequestBody{ServerId: serverID})
	return err
}

// DetachVolume detaches the volume.
func (c *Client) DetachVolume(ctx context.Context, volumeID string, force bool) error {
	body := gen.VolumeDetachJSONRequestBody{}
	if force {
		body.Force = &force
	}
	_, err := c.gen.VolumeDetachWithResponse(ctx, volumeID, body)
	return err
}

// ExtendVolume increases the volume size to newSize (Gb).
func (c *Client) ExtendVolume(ctx context.Context, id string, newSize int) error {
	_, err := c.gen.VolumeExtendWithResponse(ctx, id, gen.VolumeExtendJSONRequestBody{NewSize: newSize})
	return err
}

// DeleteVolume deletes the volume. The v3 API requires a (here empty) body.
func (c *Client) DeleteVolume(ctx context.Context, id string) error {
	_, err := c.gen.VolumeDeleteWithResponse(ctx, id, gen.VolumeDeleteJSONRequestBody{})
	return err
}
