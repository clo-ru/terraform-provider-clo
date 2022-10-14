package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_servers "github.com/clo-ru/cloapi-go-client/services/servers"
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
			"results": {
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
						"image_id": {
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
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("results", flattenInstancesResults(resp.Results)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenInstancesResults(pr []clo_servers.ServerListItem) []interface{} {
	lpr := len(pr)
	if lpr > 0 {
		res := make([]interface{}, lpr, lpr)
		for i, p := range pr {
			ri := make(map[string]interface{})
			ri["id"] = p.ID
			ri["name"] = p.Name
			ri["image_id"] = p.Image
			ri["status"] = p.Status
			ri["recipe_id"] = p.Recipe
			ri["created_in"] = p.CreatedIn
			ri["rescue_mode"] = p.InRescue
			ri["flavor_ram"] = p.Flavor.Ram
			ri["flavor_vcpus"] = p.Flavor.Vcpus
			var diskData []interface{}
			ld := len(p.DiskData)
			if ld > 0 {
				diskData = make([]interface{}, ld)
				for j, d := range p.DiskData {
					dd := map[string]interface{}{
						"id":           d.ID,
						"storage_type": d.StorageType,
					}
					diskData[j] = dd
				}
			}
			ri["disk_data"] = diskData
			res[i] = ri
		}
		return res
	}
	return make([]interface{}, 0)
}
