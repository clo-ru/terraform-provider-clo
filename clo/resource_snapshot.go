package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Server snapshot lifecycle statuses, per the cloud_image model (snapshots are
// stored as private images of type SNAPSHOT). ERROR and UNKNOWN are
// intentionally left out of every waiter's pending set, so StateChangeConf
// surfaces them as errors instead of hanging.
const (
	creatingSnapshot   = "CREATING"
	activeSnapshot     = "ACTIVE"
	processingSnapshot = "PROCESSING"
	deletingSnapshot   = "DELETING"
	deletedSnapshot    = "DELETED"
)

func resourceSnapshot() *schema.Resource {
	return &schema.Resource{
		Description:   "Take a point-in-time snapshot of a server. Snapshots are stored as private images and auto-expire at `deleted_in`.",
		ReadContext:   resourceSnapshotRead,
		CreateContext: resourceSnapshotCreate,
		DeleteContext: resourceSnapshotDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"server_id": {
				Description: "ID of the server to snapshot",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Name of the snapshot",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"id": {
				Description: "ID of the snapshot",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Lifecycle status of the snapshot",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "Snapshot size in bytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"parent_server": {
				Description: "ID of the server the snapshot was taken from",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"child_servers": {
				Description: "IDs of servers provisioned from this snapshot",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"created_in": {
				Description: "Timestamp the snapshot was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"deleted_in": {
				Description: "Timestamp the snapshot is scheduled to be auto-deleted",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceSnapshotCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.CreateSnapshot(ctx, d.Get("server_id").(string), d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitSnapshotState(ctx, id, cli, []string{creatingSnapshot, processingSnapshot}, []string{activeSnapshot}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	return resourceSnapshotRead(ctx, d, m)
}

func resourceSnapshotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	s, err := cli.GetSnapshot(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	// server_id is a write-only input; parent_server (from detail) is its
	// read-back counterpart, so server_id is preserved in state as configured.
	fields := map[string]interface{}{
		"id":            s.ID,
		"name":          s.Name,
		"status":        s.Status,
		"size":          s.Size,
		"parent_server": s.ParentServer,
		"child_servers": s.ChildServers,
		"created_in":    s.CreatedIn,
		"deleted_in":    s.DeletedIn,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceSnapshotDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteSnapshot(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitSnapshotDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Waiters

func waitSnapshotState(ctx context.Context, id string, cli *cloapi.Client, pending, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		s, err := cli.GetSnapshot(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return s, s.Status, nil
	})
}

func waitSnapshotDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	return waitForState(ctx, timeout, []string{activeSnapshot, processingSnapshot, deletingSnapshot}, []string{deletedSnapshot}, func() (interface{}, string, error) {
		s, err := cli.GetSnapshot(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedSnapshot, nil
		}
		if err != nil {
			return nil, "", err
		}
		return s, s.Status, nil
	})
}
