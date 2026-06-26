package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceIP() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the data of an IP-address with a provided ID",
		ReadContext: dataSourceIPRead,
		Schema: map[string]*schema.Schema{
			"address_id": {
				Description: "ID of the address",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ptr": {
				Description: "PTR of the attached address",
				Type:        schema.TypeString, Computed: true},
			"type": {
				Description: "Type of the attached address. " +
					"Might be one of: FLOATING, FIXED, FREE, VROUTER.",
				Type: schema.TypeString, Computed: true},
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
	}
}

func dataSourceIPRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	addr, err := cli.GetAddress(ctx, d.Get("address_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	fields := map[string]interface{}{
		"ptr":             addr.Ptr,
		"type":            addr.Type,
		"status":          addr.Status,
		"address":         addr.Address,
		"created_in":      addr.CreatedIn,
		"is_primary":      addr.IsPrimary,
		"ddos_protection": addr.DdosProtection,
		"attached_to":     flattenAttachedTo(addr.AttachedTo),
	}
	for k, v := range fields {
		if e := d.Set(k, v); e != nil {
			return diag.FromErr(e)
		}
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenAttachedTo(attach *cloapi.AddressAttachedTo) []interface{} {
	att := make([]interface{}, 0)
	if attach != nil {
		att = append(att, map[string]interface{}{"id": attach.ID, "entity": attach.Entity})
	}
	return att
}
