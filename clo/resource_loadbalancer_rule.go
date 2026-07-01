package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Listener-rule (pool member) statuses, per the cloud_pool_member schema. The
// failure statuses CREATING_ERROR and ERROR are intentionally not enumerated:
// they are absent from every waiter's pending set, so StateChangeConf surfaces
// them as errors instead of hanging.
const (
	creatingRule = "CREATING"
	activeRule   = "ACTIVE"
	deletingRule = "DELETING"
	deletedRule  = "DELETED"
)

func resourceLoadBalancerRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage a listener rule on a load balancer (maps an external port to an internal port on the backend).",
		ReadContext:   resourceLoadBalancerRuleRead,
		CreateContext: resourceLoadBalancerRuleCreate,
		DeleteContext: resourceLoadBalancerRuleDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"loadbalancer_id": {
				Description: "ID of the load balancer the rule belongs to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"address_id": {
				Description: "ID of the address the rule listens on",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"external_protocol_port": {
				Description: "Port exposed on the load balancer's address",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"internal_protocol_port": {
				Description: "Port the traffic is forwarded to on the backend",
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
			},
			"id": {
				Description: "ID of the rule",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {
				Description: "Lifecycle status of the rule",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"server": {
				Description: "ID of the backend server the rule targets",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"address": {
				Description: "Address the rule listens on",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceLoadBalancerRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.CreateRule(ctx, d.Get("loadbalancer_id").(string), cloapi.RuleCreateParams{
		AddressID:            d.Get("address_id").(string),
		ExternalProtocolPort: d.Get("external_protocol_port").(int),
		InternalProtocolPort: d.Get("internal_protocol_port").(int),
	})
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitRuleState(ctx, id, cli, []string{creatingRule}, []string{activeRule}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	return resourceLoadBalancerRuleRead(ctx, d, m)
}

func resourceLoadBalancerRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	r, err := cli.GetRule(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	fields := map[string]interface{}{
		"id":                     r.ID,
		"status":                 r.Status,
		"server":                 r.Server,
		"address":                r.Address,
		"external_protocol_port": r.ExternalProtocolPort,
		"internal_protocol_port": r.InternalProtocolPort,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceLoadBalancerRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteRule(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitRuleDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Waiters

func waitRuleState(ctx context.Context, id string, cli *cloapi.Client, pending, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		r, err := cli.GetRule(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return r, r.Status, nil
	})
}

func waitRuleDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	pending := []string{creatingRule, activeRule, deletingRule}
	return waitForState(ctx, timeout, pending, []string{deletedRule}, func() (interface{}, string, error) {
		r, err := cli.GetRule(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedRule, nil
		}
		if err != nil {
			return nil, "", err
		}
		return r, r.Status, nil
	})
}
