package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Virtual-router lifecycle statuses, per the CLO API docs. status is a single
// field that encodes both provisioning and power state: ACTIVE = running,
// STOPPED = powered off.
const (
	creatingVrouter = "CREATING"
	activeVrouter   = "ACTIVE"
	startingVrouter = "STARTING"
	stoppingVrouter = "STOPPING"
	stoppedVrouter  = "STOPPED"
	deletingVrouter = "DELETING"
	deletedVrouter  = "DELETED"
	errorVrouter    = "ERROR"
)

func resourceVrouter() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage a virtual router in the project. `enabled` toggles the router's power state (start/stop).",
		ReadContext:   resourceVrouterRead,
		CreateContext: resourceVrouterCreate,
		UpdateContext: resourceVrouterUpdate,
		DeleteContext: resourceVrouterDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the virtual router should be created",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "Human-readable name of the virtual router",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"private_networks": {
				Description: "IDs of private networks to attach to the router",
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"enabled": {
				Description: "Whether the router is powered on. Defaults to true.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"id": {
				Description: "ID of the virtual router",
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
		},
	}
}

func resourceVrouterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	p := cloapi.VrouterCreateParams{
		Name:            d.Get("name").(string),
		PrivateNetworks: expandStringList(d.Get("private_networks").([]interface{})),
	}
	id, err := cli.CreateVrouter(ctx, d.Get("project_id").(string), p)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitVrouterState(ctx, id, cli, []string{creatingVrouter, startingVrouter}, []string{activeVrouter}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	// A freshly created router comes up running; only act if the user asked for it stopped.
	if !d.Get("enabled").(bool) {
		if err := cli.StopVrouter(ctx, id); err != nil {
			return diag.FromErr(err)
		}
		if err := waitVrouterEnabled(ctx, id, cli, false, d.Timeout(schema.TimeoutCreate)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceVrouterRead(ctx, d, m)
}

func resourceVrouterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	v, err := cli.GetVrouter(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	fields := map[string]interface{}{
		"id":                          v.ID,
		"name":                        v.Name,
		"project_id":                  v.Project,
		"status":                      v.Status,
		"switch_status":               v.SwitchStatus,
		"external_gateway_address_id": v.ExternalGatewayAddressID,
		"enabled":                     vrouterEnabled(v.Status),
		"private_networks":            v.PrivateNetworks,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceVrouterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if d.HasChange("enabled") {
		enabled := d.Get("enabled").(bool)
		if enabled {
			if err := cli.StartVrouter(ctx, d.Id()); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := cli.StopVrouter(ctx, d.Id()); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := waitVrouterEnabled(ctx, d.Id(), cli, enabled, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceVrouterRead(ctx, d, m)
}

func resourceVrouterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteVrouter(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitVrouterDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// vrouterEnabled maps the lifecycle status to the power-on bool surfaced as `enabled`.
func vrouterEnabled(status string) bool {
	switch status {
	case stoppedVrouter, stoppingVrouter:
		return false
	default:
		return true
	}
}

// Waiters

func waitVrouterState(ctx context.Context, id string, cli *cloapi.Client, pending, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		v, err := cli.GetVrouter(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return v, v.Status, nil
	})
}

// waitVrouterEnabled waits for the router to settle in the running (ACTIVE) or
// stopped (STOPPED) state after a Start/Stop. An ERROR status is not in the
// pending set, so StateChangeConf surfaces it as a failure instead of hanging.
func waitVrouterEnabled(ctx context.Context, id string, cli *cloapi.Client, enabled bool, timeout time.Duration) error {
	if enabled {
		return waitVrouterState(ctx, id, cli, []string{stoppedVrouter, startingVrouter}, []string{activeVrouter}, timeout)
	}
	return waitVrouterState(ctx, id, cli, []string{activeVrouter, stoppingVrouter}, []string{stoppedVrouter}, timeout)
}

func waitVrouterDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	pending := []string{creatingVrouter, activeVrouter, startingVrouter, stoppingVrouter, stoppedVrouter, deletingVrouter}
	return waitForState(ctx, timeout, pending, []string{deletedVrouter}, func() (interface{}, string, error) {
		v, err := cli.GetVrouter(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedVrouter, nil
		}
		if err != nil {
			return nil, "", err
		}
		return v, v.Status, nil
	})
}

// expandStringList converts a schema TypeList of strings into a []string.
func expandStringList(in []interface{}) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, v := range in {
		out = append(out, v.(string))
	}
	return out
}
