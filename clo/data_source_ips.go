package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_ip "github.com/clo-ru/cloapi-go-client/v2/services/ip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceIPs() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of the IP-addresses",
		ReadContext: dataSourceIPsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns addresses",
				Type:        schema.TypeString,
				Required:    true,
			},
			"results": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the address",
							Type:        schema.TypeString, Computed: true},
						"ptr": {
							Description: "PTR of the address",
							Type:        schema.TypeString, Computed: true},
						"type": {
							Description: "Type of the attached address. Might be one of: FLOATING, FIXED, FREE, VROUTER.",
							Type:        schema.TypeString, Computed: true},
						"status": {Type: schema.TypeString, Computed: true},
						"address": {
							Description: "String representation of the address",
							Type:        schema.TypeString, Computed: true},
						"created_in": {
							Description: "Timestamp the address was created",
							Type:        schema.TypeString, Computed: true},
						"is_primary": {
							Description: "Is the IP-address using as primary",
							Type:        schema.TypeBool, Computed: true},
						"ddos_protection": {
							Description: "Is the address protected from DDoS",
							Type:        schema.TypeBool, Computed: true},
						"attached_to": {
							Description: "Information about where the address is attached",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"id": {
									Description: "ID of the object to which the address is attached",
									Type:        schema.TypeString, Computed: true},
								"entity": {
									Description: "Type of the object to which the address is attached. " +
										"For example: `loadbalancer` or `server`",
									Type: schema.TypeString, Computed: true},
							}},
						},
					},
				},
			},
		},
	}
}

func dataSourceIPsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressListRequest{
		ProjectID: d.Get("project_id").(string),
	}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("results", flattenIpsResults(resp.Result)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenIpsResults(pr []clo_ip.Address) []interface{} {
	lpr := len(pr)
	if lpr > 0 {
		res := make([]interface{}, lpr, lpr)
		for i, p := range pr {
			ri := make(map[string]interface{})
			ri["id"] = p.ID
			ri["ptr"] = p.Ptr
			ri["type"] = p.Type
			ri["status"] = p.Status
			ri["address"] = p.Address
			ri["created_in"] = p.CreatedIn
			ri["is_primary"] = p.IsPrimary
			ri["ddos_protection"] = p.DdosProtection
			ri["attached_to"] = formatAttachedTo(p.AttachedTo)
			res[i] = ri
		}
		return res
	}
	return make([]interface{}, 0)
}
