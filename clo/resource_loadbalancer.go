package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Load-balancer lifecycle statuses, per the cloud_loadbalancer schema. status
// spans provisioning, power and config transitions; switch_status is the ON/OFF
// power state that drives `enabled`. The failure statuses CREATING_ERROR and
// ERROR are intentionally not enumerated: they are absent from every waiter's
// pending set, so StateChangeConf surfaces them as errors instead of hanging.
const (
	creatingLB = "CREATING"
	activeLB   = "ACTIVE"
	startingLB = "STARTING"
	stoppingLB = "STOPPING"
	stoppedLB  = "STOPPED"
	changingLB = "CHANGING"
	updatingLB = "UPDATING"
	deletingLB = "DELETING"
	deletedLB  = "DELETED"

	switchOnLB = "ON"
)

func resourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage a load balancer in the project. `enabled` toggles the balancer's power state (start/stop).",
		ReadContext:   resourceLoadBalancerRead,
		CreateContext: resourceLoadBalancerCreate,
		UpdateContext: resourceLoadBalancerUpdate,
		DeleteContext: resourceLoadBalancerDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the load balancer should be created",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human-readable name of the load balancer",
				Type:        schema.TypeString,
				Required:    true,
			},
			"algorithm": {
				Description: "Balancing algorithm. One of `ROUND_ROBIN`, `LEAST_CONNECTIONS`",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"session_persistence": {
				Description: "Whether to keep a client on the same backend across requests",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"address": {
				Description: "Address to attach to the load balancer. If omitted, one is allocated automatically",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "Use an existing address with this ID",
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
						},
						"ddos_protection": {
							Description: "Whether the allocated address should be DDoS-protected",
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
						},
					},
				},
			},
			"healthmonitor": {
				Description: "Health-check configuration for the backend pool",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "Health-check type. One of `PING`, `TCP`, `HTTP`",
							Type:        schema.TypeString,
							Required:    true,
						},
						"delay": {
							Description: "Seconds between health checks (interval; minimum 80)",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"timeout": {
							Description: "Seconds to wait for a health-check response (minimum 15)",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"max_retries": {
							Description: "Failed checks before a backend is marked down (1-10)",
							Type:        schema.TypeInt,
							Required:    true,
						},
						"http_method": {
							Description: "HTTP method for the check (HTTP type only), e.g. `GET`",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"url_path": {
							Description: "URL path to check (HTTP type only), e.g. `/health`",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"expected_codes": {
							Description: "Expected HTTP status codes (HTTP type only), e.g. `200`",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"enabled": {
				Description: "Whether the load balancer is powered on. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"id": {
				Description: "ID of the load balancer",
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
			"addresses": {
				Description: "IDs of the addresses bound to the load balancer",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"rules_count": {
				Description: "Number of listener rules on the load balancer",
				Type:        schema.TypeInt,
				Computed:    true,
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
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := cli.CreateLoadBalancer(ctx, d.Get("project_id").(string), buildLoadBalancerCreateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitLoadBalancerState(ctx, id, cli, []string{creatingLB, startingLB}, []string{activeLB}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	// A freshly created balancer comes up running; only act if the user asked for it stopped.
	if !d.Get("enabled").(bool) {
		if err := cli.StopLoadBalancer(ctx, id); err != nil {
			return diag.FromErr(err)
		}
		if err := waitLoadBalancerEnabled(ctx, id, cli, false, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceLoadBalancerRead(ctx, d, m)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	lb, err := cli.GetLoadBalancer(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	fields := map[string]interface{}{
		"id":                  lb.ID,
		"name":                lb.Name,
		"project_id":          lb.Project,
		"status":              lb.Status,
		"switch_status":       lb.SwitchStatus,
		"algorithm":           lb.Algorithm,
		"session_persistence": lb.SessionPersistence,
		"rules_count":         lb.RulesCount,
		"enabled":             lb.SwitchStatus == switchOnLB,
		"addresses":           lb.Addresses,
		"healthmonitor":       flattenHealthmonitor(lb.Healthmonitor),
		"created_in":          lb.CreatedIn,
		"updated_in":          lb.UpdatedIn,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id := d.Id()

	changed := false
	if d.HasChange("name") {
		if err := cli.RenameLoadBalancer(ctx, id, d.Get("name").(string)); err != nil {
			return diag.FromErr(err)
		}
		changed = true
	}
	if d.HasChanges("algorithm", "session_persistence") {
		sp := d.Get("session_persistence").(bool)
		if err := cli.UpdateLoadBalancer(ctx, id, d.Get("algorithm").(string), &sp); err != nil {
			return diag.FromErr(err)
		}
		changed = true
	}
	if d.HasChange("healthmonitor") {
		if err := cli.UpdateHealthmonitor(ctx, id, expandHealthmonitor(d.Get("healthmonitor").([]interface{}))); err != nil {
			return diag.FromErr(err)
		}
		changed = true
	}
	if changed {
		if err := waitLoadBalancerSettled(ctx, id, cli, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		if enabled {
			if err := cli.EnableLoadBalancer(ctx, id); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := cli.StopLoadBalancer(ctx, id); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := waitLoadBalancerEnabled(ctx, id, cli, enabled, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceLoadBalancerRead(ctx, d, m)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteLoadBalancer(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitLoadBalancerDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func buildLoadBalancerCreateParams(d *schema.ResourceData) cloapi.LoadBalancerCreateParams {
	sp := d.Get("session_persistence").(bool)
	p := cloapi.LoadBalancerCreateParams{
		Name:               d.Get("name").(string),
		Algorithm:          d.Get("algorithm").(string),
		SessionPersistence: &sp,
		Healthmonitor:      expandHealthmonitor(d.Get("healthmonitor").([]interface{})),
	}
	if v, ok := d.GetOk("address"); ok {
		list := v.([]interface{})
		if len(list) > 0 && list[0] != nil {
			m := list[0].(map[string]interface{})
			id, _ := m["id"].(string)
			p.AddressID = id
			// ddos_protection applies only when allocating a new address; the API
			// rejects it alongside an existing address id.
			if id == "" {
				if ddos, ok := m["ddos_protection"].(bool); ok {
					dd := ddos
					p.AddressDdos = &dd
				}
			}
		}
	}
	return p
}

func expandHealthmonitor(list []interface{}) cloapi.HealthmonitorParams {
	hm := cloapi.HealthmonitorParams{}
	if len(list) == 0 || list[0] == nil {
		return hm
	}
	m := list[0].(map[string]interface{})
	if x, ok := m["type"].(string); ok {
		hm.Type = x
	}
	if x, ok := m["delay"].(int); ok {
		hm.Delay = x
	}
	if x, ok := m["timeout"].(int); ok {
		hm.Timeout = x
	}
	if x, ok := m["max_retries"].(int); ok {
		hm.MaxRetries = x
	}
	if x, ok := m["http_method"].(string); ok {
		hm.HttpMethod = x
	}
	if x, ok := m["url_path"].(string); ok {
		hm.UrlPath = x
	}
	if x, ok := m["expected_codes"].(string); ok {
		hm.ExpectedCodes = x
	}
	return hm
}

func flattenHealthmonitor(hm cloapi.Healthmonitor) []interface{} {
	return []interface{}{map[string]interface{}{
		"type":           hm.Type,
		"delay":          hm.Delay,
		"timeout":        hm.Timeout,
		"max_retries":    hm.MaxRetries,
		"http_method":    hm.HttpMethod,
		"url_path":       hm.UrlPath,
		"expected_codes": hm.ExpectedCodes,
	}}
}

// Waiters

func waitLoadBalancerState(ctx context.Context, id string, cli *cloapi.Client, pending, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		lb, err := cli.GetLoadBalancer(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return lb, lb.Status, nil
	})
}

// waitLoadBalancerEnabled waits for the balancer to settle running (ACTIVE) or
// stopped (STOPPED) after an Enable/Stop. Error statuses are not in the pending
// set, so StateChangeConf surfaces them as a failure instead of hanging.
func waitLoadBalancerEnabled(ctx context.Context, id string, cli *cloapi.Client, enabled bool, timeout time.Duration) error {
	if enabled {
		return waitLoadBalancerState(ctx, id, cli, []string{stoppedLB, startingLB}, []string{activeLB}, timeout)
	}
	return waitLoadBalancerState(ctx, id, cli, []string{activeLB, stoppingLB}, []string{stoppedLB}, timeout)
}

// waitLoadBalancerSettled waits for a config change (rename/algorithm/healthmonitor)
// to leave the transient CHANGING/UPDATING states and settle back to a steady state.
func waitLoadBalancerSettled(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	return waitLoadBalancerState(ctx, id, cli, []string{changingLB, updatingLB}, []string{activeLB, stoppedLB}, timeout)
}

func waitLoadBalancerDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	pending := []string{creatingLB, activeLB, startingLB, stoppingLB, stoppedLB, changingLB, updatingLB, deletingLB}
	return waitForState(ctx, timeout, pending, []string{deletedLB}, func() (interface{}, string, error) {
		lb, err := cli.GetLoadBalancer(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedLB, nil
		}
		if err != nil {
			return nil, "", err
		}
		return lb, lb.Status, nil
	})
}
