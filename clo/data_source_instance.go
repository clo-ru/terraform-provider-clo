package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_servers "github.com/clo-ru/cloapi-go-client/services/servers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceInstance() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the data of an instance with a provided ID",
		ReadContext: dataSourceInstanceRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the instance",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name of the instance",
				Type:        schema.TypeString, Computed: true},
			"image_id": {
				Description: "ID of the image from which the instance was installed",
				Type:        schema.TypeString, Computed: true},
			"recipe_id": {
				Description: "ID of the recipe that was installed on the instance",
				Type:        schema.TypeString, Computed: true},
			"status": {
				Description: "Current status of the instance",
				Type:        schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the instance was created",
				Type:        schema.TypeString, Computed: true},
			"project_id": {
				Description: "ID of the project that owns instances",
				Type:        schema.TypeString, Computed: true},
			"rescue_mode": {
				Description: "Describes is the instance in rescue mode",
				Type:        schema.TypeString, Computed: true},
			"flavor_vcpus": {
				Description: "Number of VCPU of the instance",
				Type:        schema.TypeInt, Computed: true},
			"flavor_ram": {
				Description: "Amount of RAM of the instance",
				Type:        schema.TypeInt, Computed: true},
			"guest_agent": {
				Description: "Is guest agent installed on the instance",
				Type:        schema.TypeBool, Computed: true},
			"disk_data": {
				Description: "Information about disks attached to the instance",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id": {
						Description: "ID of the attached disk",
						Type:        schema.TypeString, Computed: true},
					"storage_type": {
						Description: "Storage type of the attached disk. Could be `volume` or `local`",
						Type:        schema.TypeString, Computed: true},
				}}},
			"addresses": {
				Description: "Information about addresses attached to the instance",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"id": {
						Description: "ID of the attached address",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"ptr": {
						Description: "PTR of the attached address",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"name": {
						Description: "Name of the attached address",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"type": {
						Description: "Type of the attached address. " +
							"Might be one of: FLOATING, FIXED, FREE, VROUTER.",
						Type:     schema.TypeString,
						Computed: true,
					},
					"macaddr": {
						Description: "Mac-address of the attached address",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"version": {
						Description: "Version of the attached address. Could be `4` or `6`",
						Type:        schema.TypeInt,
						Computed:    true,
					},
					"external": {
						Description: "Is the attached address the external one",
						Type:        schema.TypeBool,
						Computed:    true,
					},
					"ddos_protection": {
						Description: "Is the attached address protected from DDoS",
						Type:        schema.TypeBool,
						Computed:    true,
					},
				}}},
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_servers.ServerDetailRequest{
		ServerID: d.Get("id").(string),
	}
	resp, e := req.Make(ctx, cli)
	if resp.Code == 404 {
		e = fmt.Errorf("NotFound returned")
	}
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("name", resp.Result.Name); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("image_id", resp.Result.Image); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("recipe_id", resp.Result.Recipe); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("project_id", resp.Result.ProjectID); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", resp.Result.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("guest_agent", resp.Result.GuestAgent); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("rescue_mode", resp.Result.RescueMode); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("flavor_ram", resp.Result.Flavor.Ram); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("flavor_vcpus", resp.Result.Flavor.Vcpus); e != nil {
		return diag.FromErr(e)
	}
	var diskData []interface{}
	ld := len(resp.Result.DiskData)
	if ld > 0 {
		diskData = make([]interface{}, ld)
		for j, d := range resp.Result.DiskData {
			dd := map[string]interface{}{
				"id":           d.ID,
				"storage_type": d.StorageType,
			}
			diskData[j] = dd
		}
	}
	if e := d.Set("disk_data", diskData); e != nil {
		return diag.FromErr(e)
	}
	var addressData []interface{}
	la := len(resp.Result.Addresses)
	if la > 0 {
		addressData = make([]interface{}, la)
		for j, a := range resp.Result.Addresses {
			ad := map[string]interface{}{
				"id":              a.ID,
				"ptr":             a.Ptr,
				"name":            a.Name,
				"type":            a.Type,
				"version":         a.Version,
				"macaddr":         a.MacAddr,
				"external":        a.External,
				"ddos_protection": a.DdosProtection,
			}
			addressData[j] = ad
		}
	}
	if e := d.Set("addresses", addressData); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}
