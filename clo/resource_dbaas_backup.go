package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Dbaas backup lifecycle statuses, per the cloud_dbaas_backup model. ERROR is
// intentionally left out of every waiter's pending set, so StateChangeConf
// surfaces it as an error instead of hanging.
const (
	buildBackup     = "BUILD"
	availableBackup = "AVAILABLE"
	deletingBackup  = "DELETING"
	deletedBackup   = "DELETED"
)

func resourceDbaasBackup() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a backup of a managed-database (dbaas) cluster or a single database. Set `cluster_id` for a FULL backup or `database_id` for a PARTIAL backup (exactly one is required).",
		ReadContext:   resourceDbaasBackupRead,
		CreateContext: resourceDbaasBackupCreate,
		UpdateContext: resourceDbaasBackupUpdate,
		DeleteContext: resourceDbaasBackupDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description:  "ID of the cluster to back up. Produces a FULL backup. Exactly one of `cluster_id` or `database_id` is required.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"cluster_id", "database_id"},
			},
			"database_id": {
				Description:  "ID of the database to back up. Produces a PARTIAL backup. Exactly one of `cluster_id` or `database_id` is required.",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"cluster_id", "database_id"},
			},
			"name": {
				Description: "Name of the backup. If omitted, the server assigns one.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"force_delete": {
				Description: "Delete the backup even when it is the parent of other (incremental) backups. Applied at destroy time.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"id": {
				Description: "ID of the backup",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project": {
				Description: "ID of the project the backup belongs to",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Lifecycle status of the backup",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description: "Backup type (`FULL`, `PARTIAL` or `INCREMENTAL`)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"size": {
				Description: "Total backup size in bytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"data_size": {
				Description: "Backed-up data size in bytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"parent": {
				Description: "ID of the parent backup, for incremental backups",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datastore_name": {
				Description: "Database engine of the backup's datastore",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datastore_version": {
				Description: "Database engine version of the backup's datastore",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_in": {
				Description: "Timestamp the backup was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDbaasBackupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	name := d.Get("name").(string)

	var (
		id  string
		err error
	)
	if v, ok := d.GetOk("cluster_id"); ok {
		id, err = cli.CreateClusterBackup(ctx, v.(string), name)
	} else {
		id, err = cli.CreateDatabaseBackup(ctx, d.Get("database_id").(string), name)
	}
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitBackupState(ctx, id, cli, []string{buildBackup}, []string{availableBackup}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	return resourceDbaasBackupRead(ctx, d, m)
}

func resourceDbaasBackupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	b, err := cli.GetBackup(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	// cluster_id / database_id are write-only inputs: the detail only ever
	// reports the parent cluster, so setting them here would drift a PARTIAL
	// backup's config. The configured values are preserved in state as-is.
	fields := map[string]interface{}{
		"id":                b.ID,
		"name":              b.Name,
		"project":           b.Project,
		"status":            b.Status,
		"type":              b.Type,
		"size":              b.Size,
		"data_size":         b.DataSize,
		"parent":            b.Parent,
		"datastore_name":    b.DatastoreName,
		"datastore_version": b.DatastoreVersion,
		"created_in":        b.CreatedIn,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

// resourceDbaasBackupUpdate handles changes to force_delete only (every other
// field is ForceNew); it takes no API call and just refreshes state.
func resourceDbaasBackupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDbaasBackupRead(ctx, d, m)
}

func resourceDbaasBackupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteBackup(ctx, d.Id(), d.Get("force_delete").(bool)); err != nil {
		return diag.FromErr(err)
	}
	if err := waitBackupDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Waiters

func waitBackupState(ctx context.Context, id string, cli *cloapi.Client, pending, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		b, err := cli.GetBackup(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return b, b.Status, nil
	})
}

func waitBackupDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	return waitForState(ctx, timeout, []string{availableBackup, deletingBackup}, []string{deletedBackup}, func() (interface{}, string, error) {
		b, err := cli.GetBackup(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedBackup, nil
		}
		if err != nil {
			return nil, "", err
		}
		return b, b.Status, nil
	})
}
