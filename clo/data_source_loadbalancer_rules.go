package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceLoadBalancerRules() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of listener rules on a load balancer",
		ReadContext: dataSourceLoadBalancerRulesRead,
		Schema: map[string]*schema.Schema{
			"loadbalancer_id": {
				Description: "ID of the load balancer that owns the rules",
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
							Description: "ID of the rule",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"loadbalancer": {
							Description: "ID of the load balancer the rule belongs to",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"address": {
							Description: "Address the rule listens on",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"server": {
							Description: "ID of the backend server the rule targets",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"status": {
							Description: "Lifecycle status of the rule",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"external_protocol_port": {
							Description: "Port exposed on the load balancer's address",
							Type:        schema.TypeInt,
							Computed:    true,
						},
						"internal_protocol_port": {
							Description: "Port the traffic is forwarded to on the backend",
							Type:        schema.TypeInt,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceLoadBalancerRulesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	rules, err := cli.ListRules(ctx, d.Get("loadbalancer_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("result", flattenLoadBalancerRulesResults(rules)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenLoadBalancerRulesResults(rules []cloapi.Rule) []interface{} {
	res := make([]interface{}, 0, len(rules))
	for _, r := range rules {
		res = append(res, map[string]interface{}{
			"id":                     r.ID,
			"loadbalancer":           r.Loadbalancer,
			"address":                r.Address,
			"server":                 r.Server,
			"status":                 r.Status,
			"external_protocol_port": r.ExternalProtocolPort,
			"internal_protocol_port": r.InternalProtocolPort,
		})
	}
	return res
}
