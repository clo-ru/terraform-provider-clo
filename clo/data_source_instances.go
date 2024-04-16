package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_servers "github.com/clo-ru/cloapi-go-client/v2/services/servers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceInstances() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of the instances",
		ReadContext: dataSourceInstancesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns instances",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Type:        schema.TypeList,
				Description: "The object that holds the results",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the instance",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Name of the instance",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"image": {
							Description: "ID of the image from which the instance was installed",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"recipe_id": {
							Description: "ID of the recipe that was installed on the instance",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_in": {
							Description: "Timestamp the instance was created",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"rescue_mode": {
							Description: "Describes is the instance in rescue mode",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"flavor_vcpus": {
							Description: "Number of VCPU of the instance",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"flavor_ram": {
							Description: "Amount of RAM of the instance",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"addresses": {
							Description: "Information about addresses attached to the instance",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"disk_data": {
							Description: "Information about disks attached to the instance",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"id": {
									Description: "ID of the attached disk",
									Type:        schema.TypeString,
									Computed:    true,
								},
								"storage_type": {
									Description: "Storage type of the attached disk. Could be `volume` or `local`",
									Type:        schema.TypeString,
									Computed:    true,
								},
							}}},
					},
				},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_servers.ServerListRequest{
		ProjectID: d.Get("project_id").(string),
	}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("result", flattenInstancesResults(resp.Result)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenInstancesResults(pr []clo_servers.Server) []interface{} {
	lpr := len(pr)
	if lpr > 0 {
		res := make([]interface{}, lpr, lpr)
		for i, p := range pr {

			ri := make(map[string]interface{})
			ri["id"] = p.ID
			ri["name"] = p.Name
			ri["image"] = formatImageName(p.Image)
			ri["status"] = p.Status
			ri["recipe_id"] = formatRecipeName(p.Recipe)
			ri["created_in"] = p.CreatedIn
			ri["rescue_mode"] = p.RescueMode
			ri["flavor_ram"] = p.Flavor.Ram
			ri["flavor_vcpus"] = p.Flavor.Vcpus
			ri["disk_data"] = formatDiskData(p.DiskData)
			res[i] = ri
		}
		return res
	}
	return make([]interface{}, 0)
}
