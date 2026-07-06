package cloapi

import (
	"context"
	"errors"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Cluster is the provider-facing view of a dbaas cluster. Status is the
// lifecycle state; SwitchStatus is the ON/OFF power state toggled by Start/Stop.
type Cluster struct {
	ID               string
	Name             string
	Project          string
	Status           string
	SwitchStatus     string
	DatastoreID      string
	DatastoreName    string
	DatastoreVersion string
	FlavorRam        int
	FlavorVcpus      int
	FlavorDisk       int
	StorageSize      int
	StorageUsedKB    int
	SystemDiskSize   int
	NodesCount       int
	DatabasesCount   int
	ExternalAddress  string
	InternalAddress  string
	BackupEnabled    bool
	BackupHour       int
	CreatedIn        string
}

// Datastore is a dbaas engine offering (database type + version) selectable at
// cluster create time.
type Datastore struct {
	ID      string
	Name    string
	Version string
}

// ClusterConfig is the provider-facing view of a cluster's tunable
// configuration. Each map holds parameter name → value: Current is the live
// config, Default the datastore/flavor default, LastStable the last known-good
// config. Values are heterogeneous (numbers, strings, bools, and occasionally
// multi-value lists), so they are surfaced to callers as raw interface{}.
type ClusterConfig struct {
	Current    map[string]interface{}
	Default    map[string]interface{}
	LastStable map[string]interface{}
}

// Node is a single cluster member (database instance).
type Node struct {
	ID        string
	Name      string
	ClusterID string
	Project   string
	Role      string
	Status    string
	PrivateIP string
	CreatedIn string
}

func clusterFromSchema(r *gen.DbaasClusterSchema) Cluster {
	c := Cluster{
		ID:             r.Id,
		Name:           r.Name,
		Project:        r.Project,
		Status:         r.Status,
		SwitchStatus:   r.SwitchStatus,
		StorageSize:    r.StorageSize,
		StorageUsedKB:  r.StorageUsedKB,
		SystemDiskSize: r.SystemDiskSize,
		NodesCount:     r.NodesCount,
		DatabasesCount: r.DatabasesCount,
		BackupEnabled:  r.BackupEnabled,
		BackupHour:     r.BackupHour,
		CreatedIn:      r.CreatedIn.Format(time.RFC3339),
	}
	if r.Datastore != nil {
		c.DatastoreID = r.Datastore.Id
		c.DatastoreName = r.Datastore.Name
		c.DatastoreVersion = r.Datastore.Version
	}
	if r.Flavor != nil {
		c.FlavorRam = r.Flavor.Ram
		c.FlavorVcpus = r.Flavor.Vcpus
		c.FlavorDisk = r.Flavor.Disk
	}
	if r.ExternalAddress != nil {
		c.ExternalAddress = *r.ExternalAddress
	}
	if r.InternalAddress != nil {
		c.InternalAddress = *r.InternalAddress
	}
	return c
}

func datastoreFromSchema(r *gen.DatastoreSchema) Datastore {
	return Datastore{ID: r.Id, Name: r.Name, Version: r.Version}
}

func nodeFromSchema(r *gen.DbaasNodeSchema) Node {
	return Node{
		ID:        r.Id,
		Name:      r.Name,
		ClusterID: r.ClusterId,
		Project:   r.Project,
		Role:      r.Role,
		Status:    r.Status,
		PrivateIP: r.PrivateIp,
		CreatedIn: r.CreatedIn.Format(time.RFC3339),
	}
}

// ClusterCreateParams holds the inputs for creating a dbaas cluster. Databases
// are managed separately (clo_dbaas_database); the create-time content slot is
// RestoreFromBackup, which is mutually exclusive with an initial database.
type ClusterCreateParams struct {
	Name              string
	FlavorRam         int
	FlavorVcpus       int
	StorageSize       int
	DatastoreID       string // optional; empty → server default
	AddressID         string // optional; empty → address auto-allocated
	AddressDdos       *bool  // optional; only when allocating a new address
	RestoreFromBackup string // optional; backup ID to restore into the new cluster
}

// clusterCreateBody builds the create request body, sending optional fields only
// when set.
func clusterCreateBody(p ClusterCreateParams) gen.DbaasClusterCreateJSONRequestBody {
	body := gen.DbaasClusterCreateJSONRequestBody{
		Name:        p.Name,
		StorageSize: p.StorageSize,
	}
	body.Flavor.Ram = p.FlavorRam
	body.Flavor.Vcpus = p.FlavorVcpus
	if p.DatastoreID != "" {
		id := p.DatastoreID
		body.Datastore = &id
	}
	if p.RestoreFromBackup != "" {
		b := p.RestoreFromBackup
		body.Backup = &b
	}
	if p.AddressID != "" || p.AddressDdos != nil {
		body.Address = &struct {
			DdosProtection *bool   `json:"ddos_protection,omitempty"`
			Id             *string `json:"id,omitempty"`
		}{DdosProtection: p.AddressDdos}
		if v := p.AddressID; v != "" {
			body.Address.Id = &v
		}
	}
	return body
}

// CreateCluster creates a dbaas cluster in the project and returns its ID.
func (c *Client) CreateCluster(ctx context.Context, projectID string, p ClusterCreateParams) (string, error) {
	resp, err := c.gen.DbaasClusterCreateWithResponse(ctx, projectID, clusterCreateBody(p))
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty dbaas cluster create response")
	}
	return resp.OK.Result.Id, nil
}

// GetCluster returns the cluster's current detail.
func (c *Client) GetCluster(ctx context.Context, id string) (*Cluster, error) {
	resp, err := c.gen.DbaasClusterDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty dbaas cluster detail response")
	}
	cl := clusterFromSchema(resp.OK.Result)
	return &cl, nil
}

// ListClusters returns the project's dbaas clusters (single page, matching the other list adapters).
func (c *Client) ListClusters(ctx context.Context, projectID string) ([]Cluster, error) {
	resp, err := c.gen.DbaasClustersListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Cluster, 0, len(items))
	for i := range items {
		out = append(out, clusterFromSchema(&items[i]))
	}
	return out, nil
}

// ListDatastores returns the dbaas engine offerings available in the project.
func (c *Client) ListDatastores(ctx context.Context, projectID string) ([]Datastore, error) {
	resp, err := c.gen.ProjectDbaasDatastoresWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Datastore, 0, len(items))
	for i := range items {
		out = append(out, datastoreFromSchema(&items[i]))
	}
	return out, nil
}

// GetClusterConfig returns the cluster's current, default and last-stable
// configuration parameter sets.
func (c *Client) GetClusterConfig(ctx context.Context, clusterID string) (*ClusterConfig, error) {
	resp, err := c.gen.DbaasClusterConfigWithResponse(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty dbaas cluster config response")
	}
	r := resp.OK.Result
	return &ClusterConfig{
		Current:    r.Current,
		Default:    r.Default,
		LastStable: r.LastStable,
	}, nil
}

// ListNodes returns the cluster's member nodes.
func (c *Client) ListNodes(ctx context.Context, clusterID string) ([]Node, error) {
	resp, err := c.gen.ClusterDbaasNodesListWithResponse(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Node, 0, len(items))
	for i := range items {
		out = append(out, nodeFromSchema(&items[i]))
	}
	return out, nil
}

// RenameCluster changes the cluster's name.
func (c *Client) RenameCluster(ctx context.Context, id, name string) error {
	_, err := c.gen.DbaasClusterUpdateWithResponse(ctx, id, gen.DbaasClusterUpdateJSONRequestBody{Name: name})
	return err
}

// ResizeCluster changes the cluster's flavor (vcpus/ram). The API rejects values
// equal to the current flavor.
func (c *Client) ResizeCluster(ctx context.Context, id string, vcpus, ram int) error {
	_, err := c.gen.DbaasClusterResizeWithResponse(ctx, id, gen.DbaasClusterResizeJSONRequestBody{Vcpus: vcpus, Ram: ram})
	return err
}

// ResizeClusterStorage grows the cluster's storage. The API rejects a size at or
// below the current one (storage is grow-only).
func (c *Client) ResizeClusterStorage(ctx context.Context, id string, newSize int) error {
	_, err := c.gen.DbaasClusterResizeStorageWithResponse(ctx, id, gen.DbaasClusterResizeStorageJSONRequestBody{NewSize: newSize})
	return err
}

// StartCluster powers the cluster on.
func (c *Client) StartCluster(ctx context.Context, id string) error {
	_, err := c.gen.DbaasClusterStartWithResponse(ctx, id)
	return err
}

// StopCluster powers the cluster off.
func (c *Client) StopCluster(ctx context.Context, id string) error {
	_, err := c.gen.DbaasClusterStopWithResponse(ctx, id)
	return err
}

// DeleteCluster deletes a cluster.
func (c *Client) DeleteCluster(ctx context.Context, id string) error {
	_, err := c.gen.DbaasClusterDeleteWithResponse(ctx, id)
	return err
}
