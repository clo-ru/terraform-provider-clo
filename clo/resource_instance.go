package clo

import (
	"context"
	"errors"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_tools "github.com/clo-ru/cloapi-go-client/v2/clo/request_tools"
	clo_servers "github.com/clo-ru/cloapi-go-client/v2/services/servers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
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
	cli := m.(*clo_lib.ApiClient)
	pid := d.Get("project_id").(string)

	req := clo_servers.ServerCreateRequest{
		Request:   clo_lib.Request{},
		ProjectID: pid,
		Body:      buildServerCreateBody(d),
	}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}

	d.SetId(resp.Result.ID)

	if err := waitInstanceState(ctx, resp.Result.ID, cli, []string{creatingInstance}, []string{activeInstance}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(e)
	}

	if e != nil {
		return diag.FromErr(e)
	}

	if p, ok := d.GetOk("password"); ok {
		if e := resetServerPassword(ctx, resp.Result.ID, p.(string), cli); e != nil {
			return diag.FromErr(e)
		}
	}
	return resourceInstanceRead(ctx, d, m)
}
func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
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
		if err := resetServerPassword(ctx, servID, c.(string), cli); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	servID := d.Id()

	resp, err := getServerDetails(ctx, servID, cli)
	if err != nil {
		return diag.FromErr(err)
	}

	if e := d.Set("id", resp.Result.ID); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", resp.Result.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	return nil
}
func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	servID := d.Id()
	cli := m.(*clo_lib.ApiClient)

	r, err := getServerDetails(ctx, servID, cli)
	if err != nil {
		return diag.FromErr(err)
	}

	req := clo_servers.ServerDeleteRequest{
		ServerID: servID,
		Body:     buildServerDeleteBody(&r.Result),
	}

	if err := req.Do(ctx, cli); err != nil {
		return diag.FromErr(err)
	}

	if err := waitInstanceDeleted(ctx, servID, cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Helpers

func buildServerDeleteBody(server *clo_servers.Server) clo_servers.ServerDeleteBody {
	b := clo_servers.ServerDeleteBody{
		DeleteAddresses: server.Addresses,
	}

	for _, d := range server.DiskData {
		if d.StorageType == "volume" {
			b.DeleteVolumes = append(b.DeleteVolumes, d.ID)
		}
	}
	return b
}

func buildServerCreateBody(d *schema.ResourceData) clo_servers.ServerCreateBody {
	return clo_servers.ServerCreateBody{
		Name:  d.Get("name").(string),
		Image: d.Get("image_id").(string),
		Flavor: clo_servers.ServerFlavorBody{
			Ram:   d.Get("flavor_ram").(int),
			Vcpus: d.Get("flavor_vcpus").(int),
		},
		Recipe:    buildInstanceRecipeBody(d),
		Storages:  buildInstanceStorageBody(d),
		Addresses: buildInstanceAddrBody(d),
		Licenses:  buildInstanceLicenseBody(d),
		Keypairs:  buildInstanceKeypairsBody(d),
	}
}

func buildInstanceKeypairsBody(d *schema.ResourceData) (keyPairs []string) {
	if ks, ok := d.GetOk("keypairs"); ok {
		kp := ks.([]interface{})
		keyPairs = make([]string, len(kp))
		for _, v := range kp {
			keyPairs = append(keyPairs, v.(string))
		}
		return
	}
	return
}

func buildInstanceRecipeBody(d *schema.ResourceData) string {
	if rc, ok := d.GetOk("recipe_id"); ok {
		return rc.(string)
	}
	return ""
}

func buildInstanceLicenseBody(d *schema.ResourceData) (sl []clo_servers.ServerLicenseBody) {
	if v, ok := d.GetOk("licenses"); ok {
		vl := v.([]interface{})
		for _, l := range vl {
			m := l.(map[string]interface{})
			licBody := clo_servers.ServerLicenseBody{}
			if v, ok := m["addon"]; ok {
				licBody.Addon = v.(string)
			}
			if v, ok := m["name"]; ok {
				licBody.Name = v.(string)
			}
			if v, ok := m["value"]; ok {
				licBody.Value = v.(int)
			}
			sl = append(sl, licBody)
		}
		return
	}
	return
}

func buildInstanceStorageBody(d *schema.ResourceData) (ss []clo_servers.ServerStorageBody) {
	if v, ok := d.GetOk("block_device"); ok {
		vl := v.([]any)
		for _, stor := range vl {
			m := stor.(map[string]any)
			storBody := clo_servers.ServerStorageBody{
				Bootable: false,
			}
			if v, ok := m["size"]; ok {
				storBody.Size = v.(int)
			}
			if v, ok := m["bootable"]; ok {
				storBody.Bootable = v.(bool)
			}
			if v, ok := m["storage_type"]; ok {
				storBody.StorageType = v.(string)
			}
			ss = append(ss, storBody)
		}
		return
	}
	return
}

func buildInstanceAddrBody(d *schema.ResourceData) (sa []clo_servers.ServerAddressesBody) {
	if v, ok := d.GetOk("addresses"); ok {
		vl := v.([]any)
		for _, adr := range vl {
			m := adr.(map[string]any)
			a := clo_servers.ServerAddressesBody{}
			if v, ok := m["external"]; ok {
				a.External = v.(bool)
			}
			if v, ok := m["address_id"]; ok {
				a.AddressId = v.(string)
			}
			if v, ok := m["version"]; ok {
				a.Version = v.(int)
			}
			if v, ok := m["ddos_protection"]; ok {
				a.DdosProtection = v.(bool)
			}
			if v, ok := m["bandwidth"]; ok {
				a.MaxBandwidth = v.(int)
			}
			sa = append(sa, a)
		}
		return
	}
	return
}

// Waiters
func waitInstanceDeleted(ctx context.Context, serverId string, cli *clo_lib.ApiClient, timeout time.Duration) error {
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_servers.ServerDetailRequest{ServerID: serverId}
			resp, err := req.Do(ctx, cli)

			apiError := clo_tools.DefaultError{}
			resState := resp.Result.Status

			if errors.As(err, &apiError) && apiError.Code == 404 {
				resState = deletedInstance
				err = nil
			}

			return resp.Result, resState, err
		},
		Pending:    []string{deletingInstance},
		Target:     []string{deletedInstance},
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, createStateConf.Timeout, func() *resource.RetryError {
		_, err := createStateConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}

func waitInstanceState(ctx context.Context, serverId string, cli *clo_lib.ApiClient, pending []string, target []string, timeout time.Duration) error {
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_servers.ServerDetailRequest{ServerID: serverId}
			resp, err := req.Do(ctx, cli)
			return resp, resp.Result.Status, err
		},
		Pending:    pending,
		Target:     target,
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, createStateConf.Timeout, func() *resource.RetryError {
		_, err := createStateConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}

// Api actions

func getServerDetails(ctx context.Context, id string, cli *clo_lib.ApiClient) (*clo_servers.ServerDetailResponse, error) {
	req := clo_servers.ServerDetailRequest{ServerID: id}
	return req.Do(ctx, cli)
}

func resetServerPassword(ctx context.Context, id string, password string, cli *clo_lib.ApiClient) error {
	req := clo_servers.ServerChangePasswdRequest{
		ServerID: id, Body: clo_servers.ServerChangePasswdBody{Password: password},
	}
	return req.Do(ctx, cli)
}

func resizeServer(ctx context.Context, id string, vcpus int, ram int, cli *clo_lib.ApiClient, d *schema.ResourceData) error {
	req := clo_servers.ServerResizeRequest{
		Request:  clo_lib.Request{},
		ServerID: id,
		Body:     clo_servers.ServerResizeBody{Vcpus: vcpus, Ram: ram},
	}
	if err := req.Do(ctx, cli); err != nil {
		return err
	}
	if err := waitInstanceState(ctx, id, cli, []string{resizingInstance}, []string{activeInstance, stoppedInstance}, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return err
	}
	return nil
}
