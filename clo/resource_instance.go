package clo

import (
	"context"
	"log"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	activeInstance   = "ACTIVE"
	stoppedInstance  = "STOPPED"
	creatingInstance = "BUILDING"
	resizingInstance = "RESIZING"
	deletingInstance = "DELETING"
	deletedInstance  = "DELETED"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Description:   "Project compute instance",
		ReadContext:   resourceInstanceRead,
		CreateContext: resourceInstanceCreate,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the instance should be created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the new instance",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"password": {
				Description: "Password for the new instance",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"image_id": {
				Description: "ID of the image that will be using",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"flavor_ram": {
				Description: "Amount of RAM of the new instance",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"flavor_vcpus": {
				Description: "Number of VCPU of the new instance",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"block_device": {
				Description: "Disk data for the new instance",
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bootable": {
							Description: "Is the disk bootable",
							Type:        schema.TypeBool,
							Required:    true,
						},
						"storage_type": {
							Description: "Storage type of the new disk. Could be `volume` or `local`",
							Type:        schema.TypeString,
							Required:    true,
						},
						"size": {
							Description: "Requested size of the new disk",
							Type:        schema.TypeInt,
							Required:    true,
						},
					},
				},
			},
			"addresses": {
				Description: "Addresses for the new instance",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external": {
							Description: "Should the new address be the external one",
							Type:        schema.TypeBool,
							Required:    true,
							ForceNew:    true,
						},
						"version": {
							Description: "Version of the new address. Could be `4` or `6`",
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
						},
						"address_id": {
							Description: "Use an existing IP with a provided ID",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"ddos_protection": {
							Description: "Should the new address be protected from DDoS",
							Type:        schema.TypeBool,
							Required:    true,
							ForceNew:    true,
						},
						"bandwidth": {
							Description: "Max address bandwidth, must be 100 or 1024",
							Type:        schema.TypeInt,
							Optional:    true,
						},
					},
				},
			},
			"keypairs": {
				Description: "The list contains the SSH-keypairs IDs",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"recipe_id": {
				Description: "ID of the recipe that will be installed on the instance",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"licenses": {
				Description: "The list contains licences that should be ordered with the instance",
				Type:        schema.TypeList,
				ForceNew:    true,
				Optional:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"addon": {
						Type:     schema.TypeString,
						Required: true,
					},
					"name": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"value": {
						Type:     schema.TypeInt,
						Optional: true,
					},
				}},
			},
			"id": {
				Description: "ID of the created instance",
				Type:        schema.TypeString, Computed: true},
			"status": {
				Description: "Current status of the instance",
				Type:        schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the instance was created",
				Type:        schema.TypeString, Computed: true},
		},
	}
}

// Actions

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.CreateServer(ctx, buildServerCreateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	if err := waitInstanceState(ctx, id, cli, []string{creatingInstance}, []string{activeInstance}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	if p, ok := d.GetOk("password"); ok {
		if err := cli.ChangeServerPassword(ctx, id, p.(string)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	servID := d.Id()

	if d.HasChanges("flavor_ram", "flavor_vcpus") {
		_, ram := d.GetChange("flavor_ram")
		_, vcpus := d.GetChange("flavor_vcpus")
		if err := resizeServer(ctx, servID, vcpus.(int), ram.(int), cli, d); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("password") {
		_, c := d.GetChange("password")
		if err := cli.ChangeServerPassword(ctx, servID, c.(string)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	srv, err := cli.GetServer(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if e := d.Set("id", srv.ID); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", srv.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", srv.Status); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	servID := d.Id()
	cli := m.(*providerMeta).v3

	srv, err := cli.GetServer(ctx, servID)
	if err != nil {
		return diag.FromErr(err)
	}

	var deleteVolumes []string
	for _, disk := range srv.Disks {
		if disk.StorageType == "volume" {
			deleteVolumes = append(deleteVolumes, disk.ID)
		}
	}

	if err := cli.DeleteServer(ctx, servID, srv.Addresses, deleteVolumes); err != nil {
		return diag.FromErr(err)
	}

	if err := waitInstanceDeleted(ctx, servID, cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Helpers

func buildServerCreateParams(d *schema.ResourceData) cloapi.ServerCreateParams {
	return cloapi.ServerCreateParams{
		ProjectID:   d.Get("project_id").(string),
		Name:        d.Get("name").(string),
		ImageID:     d.Get("image_id").(string),
		FlavorRam:   d.Get("flavor_ram").(int),
		FlavorVcpus: d.Get("flavor_vcpus").(int),
		RecipeID:    optString(d, "recipe_id"),
		Storages:    buildInstanceStorages(d),
		Addresses:   buildInstanceAddresses(d),
		Licenses:    buildInstanceLicenses(d),
		Keypairs:    buildInstanceKeypairs(d),
	}
}

func optString(d *schema.ResourceData, key string) string {
	if v, ok := d.GetOk(key); ok {
		return v.(string)
	}
	return ""
}

func buildInstanceKeypairs(d *schema.ResourceData) []string {
	ks, ok := d.GetOk("keypairs")
	if !ok {
		return nil
	}
	kp := ks.([]interface{})
	keyPairs := make([]string, len(kp))
	for i, v := range kp {
		keyPairs[i] = v.(string)
	}
	return keyPairs
}

func buildInstanceLicenses(d *schema.ResourceData) []cloapi.ServerLicense {
	v, ok := d.GetOk("licenses")
	if !ok {
		return nil
	}
	var out []cloapi.ServerLicense
	for _, l := range v.([]interface{}) {
		m := l.(map[string]interface{})
		lic := cloapi.ServerLicense{}
		if x, ok := m["addon"]; ok {
			lic.Addon = x.(string)
		}
		if x, ok := m["name"]; ok {
			lic.Name = x.(string)
		}
		out = append(out, lic)
	}
	return out
}

func buildInstanceStorages(d *schema.ResourceData) []cloapi.ServerStorage {
	v, ok := d.GetOk("block_device")
	if !ok {
		return nil
	}
	var out []cloapi.ServerStorage
	for _, stor := range v.([]any) {
		m := stor.(map[string]any)
		s := cloapi.ServerStorage{}
		if x, ok := m["size"]; ok {
			s.Size = x.(int)
		}
		if x, ok := m["bootable"]; ok {
			s.Bootable = x.(bool)
		}
		if x, ok := m["storage_type"]; ok {
			s.StorageType = x.(string)
		}
		out = append(out, s)
	}
	return out
}

func buildInstanceAddresses(d *schema.ResourceData) []cloapi.ServerAddress {
	v, ok := d.GetOk("addresses")
	if !ok {
		return nil
	}
	var out []cloapi.ServerAddress
	for _, adr := range v.([]any) {
		m := adr.(map[string]any)
		a := cloapi.ServerAddress{}
		if x, ok := m["external"]; ok {
			a.External = x.(bool)
		}
		if x, ok := m["address_id"]; ok {
			a.AddressID = x.(string)
		}
		if x, ok := m["version"]; ok {
			a.Version = x.(int)
		}
		if x, ok := m["ddos_protection"]; ok {
			a.DdosProtection = x.(bool)
		}
		if x, ok := m["bandwidth"]; ok {
			a.Bandwidth = x.(int)
		}
		out = append(out, a)
	}
	return out
}

// Waiters
func waitInstanceDeleted(ctx context.Context, serverId string, cli *cloapi.Client, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			srv, err := cli.GetServer(ctx, serverId)
			if cloapi.IsNotFound(err) {
				return struct{}{}, deletedInstance, nil
			}
			if err != nil {
				return nil, "", err
			}
			return srv, srv.Status, nil
		},
		Pending:    []string{deletingInstance},
		Target:     []string{deletedInstance},
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, stateConf.Timeout, func() *resource.RetryError {
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}

func waitInstanceState(ctx context.Context, serverId string, cli *cloapi.Client, pending []string, target []string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			srv, err := cli.GetServer(ctx, serverId)
			if err != nil {
				return nil, "", err
			}
			return srv, srv.Status, nil
		},
		Pending:    pending,
		Target:     target,
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, stateConf.Timeout, func() *resource.RetryError {
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}

func resizeServer(ctx context.Context, id string, vcpus int, ram int, cli *cloapi.Client, d *schema.ResourceData) error {
	if err := cli.ResizeServer(ctx, id, ram, vcpus); err != nil {
		return err
	}
	return waitInstanceState(ctx, id, cli, []string{resizingInstance}, []string{activeInstance, stoppedInstance}, d.Timeout(schema.TimeoutUpdate))
}
