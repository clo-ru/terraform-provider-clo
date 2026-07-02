package cloapi

import (
	"context"
	"errors"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Snapshot is the provider-facing view of a server snapshot. A snapshot is a
// point-in-time image of a server; it auto-expires at DeletedIn.
type Snapshot struct {
	ID           string
	Name         string
	Status       string
	Size         int
	ParentServer string
	ChildServers []string
	CreatedIn    string
	DeletedIn    string
}

func snapshotFromSchema(r *gen.SnapshotSchema) Snapshot {
	return Snapshot{
		ID:           r.Id,
		Name:         r.Name,
		Status:       r.Status,
		Size:         r.Size,
		ParentServer: r.ParentServer,
		ChildServers: append([]string(nil), r.ChildServers...),
		CreatedIn:    r.CreatedIn.Format(time.RFC3339),
		DeletedIn:    r.DeletedIn.Format(time.RFC3339),
	}
}

// CreateSnapshot snapshots the server and returns the snapshot ID.
func (c *Client) CreateSnapshot(ctx context.Context, serverID, name string) (string, error) {
	resp, err := c.gen.CreateServerSnapshotWithResponse(ctx, serverID, gen.CreateServerSnapshotJSONRequestBody{Name: name})
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty server snapshot create response")
	}
	return resp.OK.Result.Id, nil
}

// GetSnapshot returns the snapshot's current detail.
func (c *Client) GetSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	resp, err := c.gen.SnapshotDetailsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty snapshot detail response")
	}
	s := snapshotFromSchema(resp.OK.Result)
	return &s, nil
}

// ListSnapshots returns the project's snapshots (single page, matching the other list adapters).
func (c *Client) ListSnapshots(ctx context.Context, projectID string) ([]Snapshot, error) {
	resp, err := c.gen.SnapshotsListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Snapshot, 0, len(items))
	for i := range items {
		out = append(out, snapshotFromSchema(&items[i]))
	}
	return out, nil
}

// DeleteSnapshot deletes a snapshot.
func (c *Client) DeleteSnapshot(ctx context.Context, id string) error {
	_, err := c.gen.SnapshotDeleteWithResponse(ctx, id)
	return err
}

// RestoreSnapshot provisions a new server named name from the snapshot and
// returns the new server's ID. The snapshot must be in the ACTIVE status.
func (c *Client) RestoreSnapshot(ctx context.Context, id, name string) (string, error) {
	resp, err := c.gen.SnapshotRestoreWithResponse(ctx, id, gen.SnapshotRestoreJSONRequestBody{Name: name})
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty snapshot restore response")
	}
	return resp.OK.Result.Id, nil
}
