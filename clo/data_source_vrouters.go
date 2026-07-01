package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVrouters() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of virtual routers in the project",
		ReadContext: dataSourceVroutersRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns virtual routers",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the virtual router",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Name of the virtual router",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "Lifecycle status of the virtual router (CREATING/ACTIVE/STARTING/STOPPING/STOPPED/DELETING/DELETED/ERROR)",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"switch_status": {
							Description: "Desired power switch position reported by the API",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"external_gateway_address_id": {
							Description: "ID of the address used as the external gateway",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"private_networks": {
							Description: "IDs of private networks attached to the router",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceVroutersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	vrouters, err := cli.ListVrouters(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(vrouters))
	for _, v := range vrouters {
		nets := make([]interface{}, 0, len(v.PrivateNetworks))
		for _, n := range v.PrivateNetworks {
			nets = append(nets, n)
		}
		res = append(res, map[string]interface{}{
			"id":                          v.ID,
			"name":                        v.Name,
			"status":                      v.Status,
			"switch_status":               v.SwitchStatus,
			"external_gateway_address_id": v.ExternalGatewayAddressID,
			"private_networks":            nets,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
