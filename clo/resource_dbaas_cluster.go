package clo

import (
	"context"
	"fmt"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Dbaas cluster lifecycle statuses, per the cloud_dbaas model. status spans
// provisioning, power and config transitions; switch_status is the ON/OFF power
// state that drives `enabled`. The failure statuses (ERROR, CREATION_ERROR, DEAD,
// CONFIG_ERROR) are intentionally not enumerated in any waiter's pending set, so
// StateChangeConf surfaces them as errors instead of hanging.
const (
	creatingCluster = "CREATING"
	activeCluster   = "ACTIVE"
	stoppedCluster  = "STOPPED"
	startingCluster = "STARTING"
	stoppingCluster = "STOPPING"
	updatingCluster = "UPDATING"
	backupCluster   = "BACKUP"
	restoreCluster  = "RESTORE"
	deletingCluster = "DELETING"
	deletedCluster  = "DELETED"

	switchOnCluster = "ON"
)

func resourceDbaasCluster() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage a managed-database (dbaas) cluster in the project. `enabled` toggles the cluster's power state (start/stop). Databases are managed with the separate `clo_dbaas_database` resource.",
		ReadContext:   resourceDbaasClusterRead,
		CreateContext: resourceDbaasClusterCreate,
		UpdateContext: resourceDbaasClusterUpdate,
		DeleteContext: resourceDbaasClusterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(40 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(40 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the cluster should be created",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human-readable name of the cluster",
				Type:        schema.TypeString,
				Required:    true,
			},
			"datastore_id": {
				Description: "ID of the datastore (database engine + version) for the cluster. Resolve it with the `clo_dbaas_datastores` data source.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"storage_size": {
				Description: "Size of the data storage in GiB. Can only be grown, never shrunk.",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"flavor": {
				Description: "Compute flavor for the cluster nodes",
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vcpus": {
							Description: "Number of virtual CPUs",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"ram": {
							Description: "Amount of RAM in GiB",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"disk": {
							Description: "System disk size in GiB reported by the API",
							Type:        schema.TypeInt,
							Computed:    true,
						},
					},
				},
			},
			"address": {
				Description: "Address to attach to the cluster. If omitted, one is allocated automatically",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Use an existing address with this ID",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"ddos_protection": {
							Description: "Whether the allocated address should be DDoS-protected",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"restore_from_backup_id": {
				Description: "ID of a backup to restore into the new cluster. Mutually exclusive with an initial database.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"enabled": {
				Description: "Whether the cluster is powered on. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"backup_enabled": {
				Description: "Whether scheduled backups are enabled for the cluster. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"id": {
				Description: "ID of the cluster",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Lifecycle status of the cluster",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"switch_status": {
				Description: "Power switch position reported by the API (`ON`/`OFF`)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datastore_name": {
				Description: "Database engine of the selected datastore",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"datastore_version": {
				Description: "Database engine version of the selected datastore",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"external_address": {
				Description: "External address of the cluster",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"internal_address": {
				Description: "Internal address of the cluster",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"nodes_count": {
				Description: "Number of nodes in the cluster",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"databases_count": {
				Description: "Number of databases in the cluster",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"storage_used_kb": {
				Description: "Storage currently used, in KiB",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"system_disk_size": {
				Description: "System disk size in GiB",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"backup_hour": {
				Description: "UTC hour at which scheduled backups run",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"created_in": {
				Description: "Timestamp the cluster was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDbaasClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.CreateCluster(ctx, d.Get("project_id").(string), buildClusterCreateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitClusterState(ctx, id, cli, []string{creatingCluster, restoreCluster, startingCluster}, []string{activeCluster}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	// A freshly created cluster comes up running; only act if the user asked for it stopped.
	if !d.Get("enabled").(bool) {
		if err := cli.StopCluster(ctx, id); err != nil {
			return diag.FromErr(err)
		}
		if err := waitClusterEnabled(ctx, id, cli, false, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Scheduled backups are on by default at creation; only act if the user asked to disable them.
	if !d.Get("backup_enabled").(bool) {
		if err := cli.DisableClusterBackup(ctx, id); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDbaasClusterRead(ctx, d, m)
}

func resourceDbaasClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	c, err := cli.GetCluster(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	fields := map[string]interface{}{
		"id":                c.ID,
		"name":              c.Name,
		"project_id":        c.Project,
		"status":            c.Status,
		"switch_status":     c.SwitchStatus,
		"datastore_id":      c.DatastoreID,
		"datastore_name":    c.DatastoreName,
		"datastore_version": c.DatastoreVersion,
		"storage_size":      c.StorageSize,
		"storage_used_kb":   c.StorageUsedKB,
		"system_disk_size":  c.SystemDiskSize,
		"nodes_count":       c.NodesCount,
		"databases_count":   c.DatabasesCount,
		"external_address":  c.ExternalAddress,
		"internal_address":  c.InternalAddress,
		"backup_hour":       c.BackupHour,
		"backup_enabled":    c.BackupEnabled,
		"created_in":        c.CreatedIn,
		"enabled":           c.SwitchStatus == switchOnCluster,
		"flavor":            flattenClusterFlavor(c),
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceDbaasClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id := d.Id()

	changed := false
	if d.HasChange("name") {
		if err := cli.RenameCluster(ctx, id, d.Get("name").(string)); err != nil {
			return diag.FromErr(err)
		}
		changed = true
	}
	if d.HasChange("flavor") {
		fl := d.Get("flavor").([]interface{})
		if len(fl) > 0 && fl[0] != nil {
			mfl := fl[0].(map[string]interface{})
			if err := cli.ResizeCluster(ctx, id, mfl["vcpus"].(int), mfl["ram"].(int)); err != nil {
				return diag.FromErr(err)
			}
			changed = true
		}
	}
	if d.HasChange("storage_size") {
		old, nw := d.GetChange("storage_size")
		if nw.(int) <= old.(int) {
			return diag.FromErr(fmt.Errorf("storage_size can only be increased (was %d, got %d)", old.(int), nw.(int)))
		}
		if err := cli.ResizeClusterStorage(ctx, id, nw.(int)); err != nil {
			return diag.FromErr(err)
		}
		changed = true
	}
	if changed {
		if err := waitClusterSettled(ctx, id, cli, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		if enabled {
			if err := cli.StartCluster(ctx, id); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := cli.StopCluster(ctx, id); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := waitClusterEnabled(ctx, id, cli, enabled, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("backup_enabled") {
		if d.Get("backup_enabled").(bool) {
			if err := cli.EnableClusterBackup(ctx, id); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := cli.DisableClusterBackup(ctx, id); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceDbaasClusterRead(ctx, d, m)
}

func resourceDbaasClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteCluster(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitClusterDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func buildClusterCreateParams(d *schema.ResourceData) cloapi.ClusterCreateParams {
	p := cloapi.ClusterCreateParams{
		Name:              d.Get("name").(string),
		StorageSize:       d.Get("storage_size").(int),
		DatastoreID:       d.Get("datastore_id").(string),
		RestoreFromBackup: d.Get("restore_from_backup_id").(string),
	}
	if fl := d.Get("flavor").([]interface{}); len(fl) > 0 && fl[0] != nil {
		mfl := fl[0].(map[string]interface{})
		p.FlavorVcpus = mfl["vcpus"].(int)
		p.FlavorRam = mfl["ram"].(int)
	}
	if v, ok := d.GetOk("address"); ok {
		list := v.([]interface{})
		if len(list) > 0 && list[0] != nil {
			ma := list[0].(map[string]interface{})
			id, _ := ma["id"].(string)
			p.AddressID = id
			// ddos_protection applies only when allocating a new address; the API
			// rejects it alongside an existing address id.
			if id == "" {
				if ddos, ok := ma["ddos_protection"].(bool); ok {
					dd := ddos
					p.AddressDdos = &dd
				}
			}
		}
	}
	return p
}

func flattenClusterFlavor(c *cloapi.Cluster) []interface{} {
	return []interface{}{map[string]interface{}{
		"vcpus": c.FlavorVcpus,
		"ram":   c.FlavorRam,
		"disk":  c.FlavorDisk,
	}}
}

// Waiters

func waitClusterState(ctx context.Context, id string, cli *cloapi.Client, pending, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		c, err := cli.GetCluster(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return c, c.Status, nil
	})
}

// waitClusterEnabled waits for the cluster to settle running (ACTIVE) or stopped
// (STOPPED) after a Start/Stop. Failure statuses are not in the pending set, so
// StateChangeConf surfaces them as a failure instead of hanging.
func waitClusterEnabled(ctx context.Context, id string, cli *cloapi.Client, enabled bool, timeout time.Duration) error {
	if enabled {
		return waitClusterState(ctx, id, cli, []string{stoppedCluster, startingCluster}, []string{activeCluster}, timeout)
	}
	return waitClusterState(ctx, id, cli, []string{activeCluster, stoppingCluster}, []string{stoppedCluster}, timeout)
}

// waitClusterSettled waits for a config change (rename/resize) to leave the
// transient UPDATING state and settle back to whatever power state it held.
func waitClusterSettled(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	return waitClusterState(ctx, id, cli, []string{updatingCluster}, []string{activeCluster, stoppedCluster}, timeout)
}

func waitClusterDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	pending := []string{creatingCluster, activeCluster, stoppedCluster, startingCluster, stoppingCluster, updatingCluster, backupCluster, restoreCluster, deletingCluster}
	return waitForState(ctx, timeout, pending, []string{deletedCluster}, func() (interface{}, string, error) {
		c, err := cli.GetCluster(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedCluster, nil
		}
		if err != nil {
			return nil, "", err
		}
		return c, c.Status, nil
	})
}
