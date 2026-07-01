package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLoadBalancers() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of load balancers in the project",
		ReadContext: dataSourceLoadBalancersRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns load balancers",
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
							Description: "ID of the load balancer",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Name of the load balancer",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "Lifecycle status of the load balancer",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"switch_status": {
							Description: "Power switch position reported by the API (`ON`/`OFF`)",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"algorithm": {
							Description: "Balancing algorithm",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"session_persistence": {
							Description: "Whether a client is kept on the same backend across requests",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"rules_count": {
							Description: "Number of listener rules on the load balancer",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"addresses": {
							Description: "IDs of the addresses bound to the load balancer",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"healthmonitor": {
							Description: "Health-check configuration for the backend pool",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"type":           {Type: schema.TypeString, Computed: true},
								"delay":          {Type: schema.TypeInt, Computed: true},
								"timeout":        {Type: schema.TypeInt, Computed: true},
								"max_retries":    {Type: schema.TypeInt, Computed: true},
								"http_method":    {Type: schema.TypeString, Computed: true},
								"url_path":       {Type: schema.TypeString, Computed: true},
								"expected_codes": {Type: schema.TypeString, Computed: true},
							}},
						},
						"created_in": {
							Description: "Timestamp the load balancer was created",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"updated_in": {
							Description: "Timestamp the load balancer was last updated",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceLoadBalancersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	lbs, err := cli.ListLoadBalancers(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("result", flattenLoadBalancersResults(lbs)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenLoadBalancersResults(lbs []cloapi.LoadBalancer) []interface{} {
	res := make([]interface{}, 0, len(lbs))
	for _, lb := range lbs {
		addrs := make([]interface{}, 0, len(lb.Addresses))
		for _, a := range lb.Addresses {
			addrs = append(addrs, a)
		}
		res = append(res, map[string]interface{}{
			"id":                  lb.ID,
			"name":                lb.Name,
			"status":              lb.Status,
			"switch_status":       lb.SwitchStatus,
			"algorithm":           lb.Algorithm,
			"session_persistence": lb.SessionPersistence,
			"rules_count":         lb.RulesCount,
			"addresses":           addrs,
			"healthmonitor":       flattenHealthmonitor(lb.Healthmonitor),
			"created_in":          lb.CreatedIn,
			"updated_in":          lb.UpdatedIn,
		})
	}
	return res
}
