package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_servers "github.com/clo-ru/cloapi-go-client/v2/services/servers"
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
			"image": {
				Description: "ID of the image from which the instance was installed",
				Type:        schema.TypeString, Computed: true},
			"recipe": {
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
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_servers.ServerDetailRequest{
		ServerID: d.Get("id").(string),
	}
	resp, e := req.Do(ctx, cli)

	if e != nil {
		return diag.FromErr(e)
	}

	if e := d.Set("name", resp.Result.Name); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("project_id", resp.Result.Project); e != nil {
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

	if e := d.Set("recipe", formatRecipeName(resp.Result.Recipe)); e != nil {
		return diag.FromErr(e)
	}

	if e := d.Set("image", formatImageName(resp.Result.Image)); e != nil {
		return diag.FromErr(e)
	}

	if e := d.Set("disk_data", formatDiskData(resp.Result.DiskData)); e != nil {
		return diag.FromErr(e)
	}

	if e := d.Set("addresses", resp.Result.Addresses); e != nil {
		return diag.FromErr(e)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func formatImageName(image *clo_servers.ServerImage) string {
	if image != nil && image.OperationSystem != nil {
		return fmt.Sprint(image.OperationSystem.Distribution, " ", image.OperationSystem.Version)
	}
	return ""
}

func formatRecipeName(recipe *clo_servers.ServerRecipe) string {
	if recipe != nil {
		return recipe.Name
	}
	return ""
}

func formatDiskData(disks []clo_servers.ServerDisk) []interface{} {
	var diskData []interface{}
	ld := len(disks)
	if ld > 0 {
		diskData = make([]interface{}, ld)
		for j, d := range disks {
			diskData[j] = map[string]interface{}{"id": d.ID, "storage_type": d.StorageType}

		}
	}
	return diskData
}
