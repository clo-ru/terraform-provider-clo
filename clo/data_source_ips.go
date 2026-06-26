package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"result": {
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
	cli := m.(*providerMeta).v3
	addresses, err := cli.ListAddresses(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("result", flattenIpsResults(addresses)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenIpsResults(pr []cloapi.Address) []interface{} {
	res := make([]interface{}, 0, len(pr))
	for _, p := range pr {
		res = append(res, map[string]interface{}{
			"id":              p.ID,
			"ptr":             p.Ptr,
			"type":            p.Type,
			"status":          p.Status,
			"address":         p.Address,
			"created_in":      p.CreatedIn,
			"is_primary":      p.IsPrimary,
			"ddos_protection": p.DdosProtection,
			"attached_to":     flattenAttachedTo(p.AttachedTo),
		})
	}
	return res
}
