package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_servers "github.com/clo-ru/cloapi-go-client/services/servers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

const (
	activeInstance   = "ACTIVE"
	creatingInstance = "BUILDING"
	resizingInstance = "RESIZING"
	deletingInstance = "DELETING"
	deletedInstance  = "DELETED"
)

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a new instance in the project",
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
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external": {
							Description: "Should the new address be the external one",
							Type:        schema.TypeBool,
							ForceNew:    true,
							Required:    true,
						},
						"version": {
							Description: "Version of the new address. Could be `4` or `6`",
							Type:        schema.TypeInt,
							Required:    true,
							ForceNew:    true,
						},
						"floatingip_id": {
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

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	pid := d.Get("project_id").(string)
	sCreateBody := clo_servers.ServerCreateBody{
		Name:  d.Get("name").(string),
		Image: d.Get("image_id").(string),
		Flavor: clo_servers.ServerFlavorBody{
			Ram:   d.Get("flavor_ram").(int),
			Vcpus: d.Get("flavor_vcpus").(int),
		},
		Storages:  buildInstanceStorageBody(d),
		Addresses: buildInstanceAddrBody(d),
	}
	if _, ok := d.GetOk("licenses"); ok {
		sCreateBody.Licenses = buildInstanceLicenseBody(d)
	}
	if rc, ok := d.GetOk("recipe_id"); ok {
		sCreateBody.Recipe = rc.(string)
	}
	if ks, ok := d.GetOk("keypairs"); ok {
		kp := ks.([]interface{})
		keyPairs := make([]string, len(kp))
		for i, v := range kp {
			keyPairs[i] = v.(string)
		}
		sCreateBody.Keypairs = keyPairs
	}
	req := clo_servers.ServerCreateRequest{
		Request:   clo_lib.Request{},
		ProjectID: pid,
		Body:      sCreateBody,
	}
	resp, e := req.Make(ctx, cli)
	if resp.Code == 404 {
		e = fmt.Errorf("NotFound returned")
	}
	if e != nil {
		return diag.FromErr(e)
	}
	d.SetId(resp.Result.ID)
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_servers.ServerDetailRequest{ServerID: resp.Result.ID}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, resp.Result.Status, e
			} else {
				return resp, resp.Result.Status, nil
			}
		},
		Pending:    []string{creatingInstance},
		Target:     []string{activeInstance},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 30 * time.Second,
	}
	e = resource.RetryContext(ctx, createStateConf.Timeout, func() *resource.RetryError {
		_, err := createStateConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
	if e != nil {
		return diag.FromErr(e)
	}
	if p, ok := d.GetOk("password"); ok {
		req := clo_servers.ServerChangePasswdRequest{
			ServerID: resp.Result.ID,
			Body:     clo_servers.ServerChangePasswdBody{Password: p.(string)},
		}
		if e := req.Make(ctx, cli); e != nil {
			return diag.FromErr(e)
		}
	}
	return resourceInstanceRead(ctx, d, m)
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
		vl := v.([]interface{})
		for _, stor := range vl {
			m := stor.(map[string]interface{})
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
		vl := v.([]interface{})
		for _, adr := range vl {
			m := adr.(map[string]interface{})
			a := clo_servers.ServerAddressesBody{}
			if v, ok := m["external"]; ok {
				a.External = v.(bool)
			}
			if v, ok := m["floatingip_id"]; ok {
				a.FloatingIpID = v.(string)
			}
			if v, ok := m["version"]; ok {
				a.Version = v.(int)
			}
			if v, ok := m["ddos_protection"]; ok {
				a.DdosProtection = v.(bool)
			}
			sa = append(sa, a)
		}
		return
	}
	return
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	servID := d.Id()
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_servers.ServerDetailRequest{ServerID: servID}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			} else {
				return resp, resp.Result.Status, nil
			}
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Minute,
	}
	if d.HasChanges("flavor_ram", "flavor_vcpus") {
		createStateConf.Target = []string{activeInstance}
		createStateConf.Pending = []string{resizingInstance}
		req := clo_servers.ServerResizeRequest{
			ServerID: servID,
			Body: clo_servers.ServerResizeBody{
				Ram:   d.Get("flavor_ram").(int),
				Vcpus: d.Get("flavor_vcpus").(int),
			},
		}
		if e := req.Make(ctx, cli); e != nil {
			return diag.FromErr(e)
		}
		_, err := createStateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if d.HasChange("password") {
		_, c := d.GetChange("password")
		req := clo_servers.ServerChangePasswdRequest{
			ServerID: servID,
			Body:     clo_servers.ServerChangePasswdBody{Password: c.(string)},
		}
		if e := req.Make(ctx, cli); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	servID := d.Id()
	req := clo_servers.ServerDetailRequest{
		ServerID: servID,
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
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
	r, e := getServerDetails(servID, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	b := clo_servers.ServerDeleteBody{}
	for _, a := range r.Addresses {
		b.DeleteAddresses = append(b.DeleteAddresses, a.ID)
	}
	for _, d := range r.DiskData {
		if d.StorageType == "volume" {
			b.DeleteVolumes = append(b.DeleteVolumes, d.ID)
		}
	}
	req := clo_servers.ServerDeleteRequest{
		ServerID: servID,
		Body:     b,
	}
	if e = req.Make(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_servers.ServerDetailRequest{ServerID: servID}
			resp, e := req.Make(ctx, cli)
			if resp.Code == 404 {
				return resp.Result, deletedInstance, nil
			}
			if e != nil {
				return resp, "", e
			}
			return resp.Result, resp.Result.Status, nil
		},
		Pending:    []string{deletingInstance},
		Target:     []string{deletedInstance},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Minute,
	}
	e = resource.RetryContext(ctx, createStateConf.Timeout, func() *resource.RetryError {
		_, err := createStateConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
	if e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func getServerDetails(id string, cli *clo_lib.ApiClient) (clo_servers.ServerDetailItem, error) {
	req := clo_servers.ServerDetailRequest{
		ServerID: id,
	}
	res, e := req.Make(context.Background(), cli)
	if e != nil {
		return clo_servers.ServerDetailItem{}, e
	}
	if res.Code == 404 {
		e = fmt.Errorf("NotFound returned")
	}
	return res.Result, nil
}
