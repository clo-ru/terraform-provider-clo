package cloapi

import (
	"context"
	"errors"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Backup is the provider-facing view of a dbaas backup. A backup is either FULL
// (created from a cluster) or PARTIAL (created from a single database).
type Backup struct {
	ID               string
	Name             string
	ClusterID        string
	Project          string
	Status           string
	Type             string
	Size             int
	DataSize         int
	Parent           string
	DatastoreName    string
	DatastoreVersion string
	CreatedIn        string
}

func backupFromSchema(r *gen.DbaasBackupSchema) Backup {
	b := Backup{
		ID:               r.Id,
		Name:             r.Name,
		ClusterID:        r.ClusterId,
		Project:          r.Project,
		Status:           r.Status,
		Type:             r.Type,
		Size:             r.Size,
		DatastoreName:    r.Datastore.Name,
		DatastoreVersion: r.Datastore.Version,
		CreatedIn:        r.CreatedIn.Format(time.RFC3339),
	}
	if r.DataSize != nil {
		b.DataSize = *r.DataSize
	}
	if r.Parent != nil {
		b.Parent = *r.Parent
	}
	return b
}

// CreateClusterBackup creates a FULL backup of the cluster and returns the backup ID.
func (c *Client) CreateClusterBackup(ctx context.Context, clusterID, name string) (string, error) {
	resp, err := c.gen.DbaasClusterBackupWithResponse(ctx, clusterID, gen.DbaasClusterBackupJSONRequestBody{Name: name})
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty dbaas cluster backup create response")
	}
	return resp.OK.Result.Id, nil
}

// CreateDatabaseBackup creates a PARTIAL backup of a single database and returns the backup ID.
func (c *Client) CreateDatabaseBackup(ctx context.Context, databaseID, name string) (string, error) {
	resp, err := c.gen.ClusterDatabaseBackupWithResponse(ctx, databaseID, gen.ClusterDatabaseBackupJSONRequestBody{Name: name})
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty dbaas database backup create response")
	}
	return resp.OK.Result.Id, nil
}

// GetBackup returns the backup's current detail.
func (c *Client) GetBackup(ctx context.Context, id string) (*Backup, error) {
	resp, err := c.gen.DbaasBackupDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty dbaas backup detail response")
	}
	b := backupFromSchema(resp.OK.Result)
	return &b, nil
}

// ListBackups returns the project's dbaas backups (single page, matching the other list adapters).
func (c *Client) ListBackups(ctx context.Context, projectID string) ([]Backup, error) {
	resp, err := c.gen.ProjectBackupListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Backup, 0, len(items))
	for i := range items {
		out = append(out, backupFromSchema(&items[i]))
	}
	return out, nil
}

// DownloadBackup returns a fresh presigned download URL for the backup.
func (c *Client) DownloadBackup(ctx context.Context, id string) (string, error) {
	resp, err := c.gen.DbaasBackupDownloadWithResponse(ctx, id)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty dbaas backup download response")
	}
	return resp.OK.Result.Url, nil
}

// DeleteBackup deletes a backup. force removes it even when it is the parent of
// other (incremental) backups.
func (c *Client) DeleteBackup(ctx context.Context, id string, force bool) error {
	body := gen.DbaasBackupDeleteJSONRequestBody{}
	if force {
		body.Force = &force
	}
	_, err := c.gen.DbaasBackupDeleteWithResponse(ctx, id, body)
	return err
}

// EnableClusterBackup turns scheduled backups on for the cluster.
func (c *Client) EnableClusterBackup(ctx context.Context, clusterID string) error {
	_, err := c.gen.DbaasClusterBackupEnableWithResponse(ctx, clusterID)
	return err
}

// DisableClusterBackup turns scheduled backups off for the cluster.
func (c *Client) DisableClusterBackup(ctx context.Context, clusterID string) error {
	_, err := c.gen.DbaasClusterBackupDisableWithResponse(ctx, clusterID)
	return err
}

// EnableDatabaseBackup turns scheduled backups on for the database.
func (c *Client) EnableDatabaseBackup(ctx context.Context, databaseID string) error {
	_, err := c.gen.DbaasDatabaseBackupEnableWithResponse(ctx, databaseID)
	return err
}

// DisableDatabaseBackup turns scheduled backups off for the database.
func (c *Client) DisableDatabaseBackup(ctx context.Context, databaseID string) error {
	_, err := c.gen.DbaasDatabaseBackupDisableWithResponse(ctx, databaseID)
	return err
}
