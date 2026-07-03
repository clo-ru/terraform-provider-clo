package clo

import (
	"context"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInstancePower() *schema.Resource {
	return &schema.Resource{
		Description: "Manage the power state of a compute instance independently of its configuration. " +
			"Power management is opt-in: without this resource the instance is left in whatever state it " +
			"boots into. Destroying this resource stops managing power — it does not change the instance's " +
			"current state or delete it.",
		ReadContext:   resourceInstancePowerRead,
		CreateContext: resourceInstancePowerCreate,
		UpdateContext: resourceInstancePowerUpdate,
		DeleteContext: resourceInstancePowerDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Description: "ID of the instance whose power state is managed",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"enabled": {
				Description: "Whether the instance should be powered on (`true`) or off (`false`)",
				Type:        schema.TypeBool,
				Required:    true,
			},
			"status": {
				Description: "Lifecycle status of the instance",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"switch_status": {
				Description: "Power switch position reported by the API (`ON`/`OFF`)",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceInstancePowerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id := d.Get("instance_id").(string)
	if err := applyServerPower(ctx, cli, id, d.Get("enabled").(bool), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return resourceInstancePowerRead(ctx, d, m)
}

func resourceInstancePowerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	srv, err := cli.GetServer(ctx, d.Id())
	if cloapi.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	// enabled is read back from the live switch_status so an out-of-band power
	// change is detected as drift and reconciled on the next apply.
	fields := map[string]interface{}{
		"instance_id":   srv.ID,
		"status":        srv.Status,
		"switch_status": srv.SwitchStatus,
		"enabled":       srv.SwitchStatus == switchOnServer,
	}
	for k, val := range fields {
		if e := d.Set(k, val); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func resourceInstancePowerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if d.HasChange("enabled") {
		if err := applyServerPower(ctx, cli, d.Id(), d.Get("enabled").(bool), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceInstancePowerRead(ctx, d, m)
}

// resourceInstancePowerDelete stops managing the instance's power; it
// deliberately leaves the instance in its current state.
func resourceInstancePowerDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

// applyServerPower drives the instance to the desired power state, doing nothing
// if it is already there (so Start on a running instance / Stop on a stopped one
// is never issued).
func applyServerPower(ctx context.Context, cli *cloapi.Client, id string, enabled bool, timeout time.Duration) error {
	srv, err := cli.GetServer(ctx, id)
	if err != nil {
		return err
	}
	if (srv.SwitchStatus == switchOnServer) == enabled {
		return nil
	}
	if enabled {
		if err := cli.StartServer(ctx, id); err != nil {
			return err
		}
	} else {
		if err := cli.StopServer(ctx, id); err != nil {
			return err
		}
	}
	return waitInstanceEnabled(ctx, id, cli, enabled, timeout)
}
