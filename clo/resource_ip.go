package clo

import (
	"context"
	"log"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	detachedIp   = "DOWN"
	attachedIp   = "ACTIVE"
	deletedIp    = "DELETED"
	processingIp = "PROCESSING"
)

func resourceIp() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a new address in the project",
		ReadContext:   resourceIpRead,
		CreateContext: resourceIpCreate,
		UpdateContext: resourceIpUpdate,
		DeleteContext: resourceIpDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "ID of the created address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"status": {Type: schema.TypeString, Computed: true},
			"is_primary": {
				Description: "Should the address be used as primary",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"bandwidth": {
				Description: "Maximum address bandwidth on mbps",
				Type:        schema.TypeInt,
				Optional:    false,
				Computed:    true,
			},
			"address": {
				Description: "String representation of the address",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ptr": {
				Description: "PTR of the attached address",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"created_in": {
				Description: "Timestamp the address was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"project_id": {
				Description: "ID of the project where the address should be created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"ddos_protection": {
				Description: "Should the address be protected from DDoS",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceIpCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3

	id, err := cli.CreateAddress(ctx, d.Get("project_id").(string), d.Get("ddos_protection").(bool))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	if err := waitAddressState(ctx, id, cli, []string{processingIp}, []string{detachedIp}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	if v, ok := d.GetOk("ptr"); ok {
		if err := cli.ChangeAddressPtr(ctx, id, v.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceIpRead(ctx, d, m)
}

func resourceIpRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3

	addr, err := cli.GetAddress(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if e := d.Set("id", addr.ID); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", addr.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("address", addr.Address); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", addr.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("bandwidth", addr.Bandwidth); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("is_primary", addr.IsPrimary); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceIpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if d.HasChange("ptr") {
		_, c := d.GetChange("ptr")
		if err := cli.ChangeAddressPtr(ctx, d.Id(), c.(string)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceIpDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteAddress(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitAddressDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Waiters
func waitAddressState(ctx context.Context, id string, cli *cloapi.Client, pending []string, target []string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			addr, err := cli.GetAddress(ctx, id)
			if err != nil {
				return nil, "", err
			}
			return addr, addr.Status, nil
		},
		Pending:    pending,
		Target:     target,
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, stateConf.Timeout, func() *resource.RetryError {
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}

func waitAddressDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			addr, err := cli.GetAddress(ctx, id)
			if cloapi.IsNotFound(err) {
				return struct{}{}, deletedIp, nil
			}
			if err != nil {
				return nil, "", err
			}
			return addr, addr.Status, nil
		},
		Pending:    []string{processingIp},
		Target:     []string{deletedIp},
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, stateConf.Timeout, func() *resource.RetryError {
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}
