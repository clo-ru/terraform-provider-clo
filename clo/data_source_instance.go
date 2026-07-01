package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	cli := m.(*providerMeta).v3
	srv, err := cli.GetServer(ctx, d.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	fields := map[string]interface{}{
		"name":         srv.Name,
		"status":       srv.Status,
		"project_id":   srv.Project,
		"created_in":   srv.CreatedIn,
		"guest_agent":  srv.GuestAgent,
		"rescue_mode":  srv.RescueMode,
		"flavor_ram":   srv.FlavorRam,
		"flavor_vcpus": srv.FlavorVcpus,
		"recipe":       srv.RecipeName,
		"image":        srv.ImageName,
		"disk_data":    flattenServerDisks(srv.Disks),
		"addresses":    srv.Addresses,
	}
	for k, v := range fields {
		if e := d.Set(k, v); e != nil {
			return diag.FromErr(e)
		}
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenServerDisks(disks []cloapi.ServerDisk) []interface{} {
	out := make([]interface{}, 0, len(disks))
	for _, dd := range disks {
		out = append(out, map[string]interface{}{"id": dd.ID, "storage_type": dd.StorageType})
	}
	return out
}
