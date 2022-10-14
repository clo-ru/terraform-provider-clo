package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_ip "github.com/clo-ru/cloapi-go-client/services/ip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
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
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressDetailRequest{
		AddressID: d.Get("address_id").(string),
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("ptr", resp.Result.Ptr); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("type", resp.Result.Type); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("type", resp.Result.Type); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("address", resp.Result.Address); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", resp.Result.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("is_primary", resp.Result.IsPrimary); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("ddos_protection", resp.Result.DdosProtection); e != nil {
		return diag.FromErr(e)
	}
	att := []interface{}{
		map[string]interface{}{
			"id":     resp.Result.AttachedTo.ID,
			"entity": resp.Result.AttachedTo.Entity,
		},
	}
	if e := d.Set("attached_to", att); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}
