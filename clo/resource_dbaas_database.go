package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Dbaas database lifecycle statuses, per the cloud_dbaas_database model. ERROR is
// intentionally left out of every waiter's pending set, so StateChangeConf
// surfaces it as an error instead of hanging.
const (
	buildDatabase    = "BUILD"
	readyDatabase    = "READY"
	deletingDatabase = "DELETING"
	deletedDatabase  = "DELETED"
)

func resourceDbaasDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage a database inside a managed-database (dbaas) cluster. Each database carries its own admin user; rotate the admin password by changing `admin_password`.",
		ReadContext:   resourceDbaasDatabaseRead,
		CreateContext: resourceDbaasDatabaseCreate,
		UpdateContext: resourceDbaasDatabaseUpdate,
		DeleteContext: resourceDbaasDatabaseDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "ID of the dbaas cluster the database belongs to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Name of the database",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"admin_username": {
				Description: "Username of the database's admin user",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"admin_password": {
				Description: "Password of the database's admin user. Changing it rotates the password on the running database. Write-only: the API never returns it, so its value is tracked from configuration.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"id": {
				Description: "ID of the database",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project": {
				Description: "ID of the project the database belongs to",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Lifecycle status of the database",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"backup_enabled": {
				Description: "Whether scheduled backups are enabled for the database",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"created_in": {
				Description: "Timestamp the database was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDbaasDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.CreateDatabase(ctx, d.Get("cluster_id").(string), cloapi.DatabaseCreateParams{
		Name:          d.Get("name").(string),
		AdminUsername: d.Get("admin_username").(string),
		AdminPassword: d.Get("admin_password").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitDatabaseState(ctx, id, cli, []string{buildDatabase}, []string{readyDatabase}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	return resourceDbaasDatabaseRead(ctx, d, m)
}

func resourceDbaasDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	db, err := cli.GetDatabase(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	// admin_password is deliberately not read back: the API never returns it, so
	// the configured value is preserved in state as-is.
	fields := map[string]interface{}{
		"id":             db.ID,
		"name":           db.Name,
		"cluster_id":     db.ClusterID,
		"project":        db.Project,
		"admin_username": db.AdminUsername,
		"status":         db.Status,
		"backup_enabled": db.BackupEnabled,
		"created_in":     db.CreatedIn,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceDbaasDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if d.HasChange("admin_password") {
		if err := cli.RestoreAdminPassword(ctx, d.Id(), d.Get("admin_password").(string)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceDbaasDatabaseRead(ctx, d, m)
}

func resourceDbaasDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteDatabase(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitDatabaseDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Waiters

func waitDatabaseState(ctx context.Context, id string, cli *cloapi.Client, pending, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		db, err := cli.GetDatabase(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return db, db.Status, nil
	})
}

func waitDatabaseDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	return waitForState(ctx, timeout, []string{readyDatabase, deletingDatabase}, []string{deletedDatabase}, func() (interface{}, string, error) {
		db, err := cli.GetDatabase(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedDatabase, nil
		}
		if err != nil {
			return nil, "", err
		}
		return db, db.Status, nil
	})
}
