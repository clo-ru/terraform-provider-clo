package clo

import (
	"context"
	"errors"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_tools "github.com/clo-ru/cloapi-go-client/v2/clo/request_tools"
	clo_ip "github.com/clo-ru/cloapi-go-client/v2/services/ip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
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
				Optional:    true,
				Computed:    true,
			},
			"bandwidth": {
				Description: "Should the address be used as primary",
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
	cli := m.(*clo_lib.ApiClient)

	req := clo_ip.AddressCreateRequest{
		ProjectID: d.Get("project_id").(string),
	}
	if n, ok := d.GetOk("ddos_protection"); ok {
		req.Body = clo_ip.AddressCreateBody{DdosProtection: n.(bool)}
	}

	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}

	d.SetId(resp.Result.ID)

	if _, err := waitAddressState(ctx, resp.Result.ID, cli, []string{processingIp}, []string{detachedIp}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(e)
	}

	if v, ok := d.GetOk("ptr"); ok {
		if err := changeAddressPtr(ctx, d, v.(string), cli); err != nil {
			return diag.FromErr(e)
		}
	}

	return resourceIpRead(ctx, d, m)
}

func resourceIpRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fipId := d.Id()
	cli := m.(*clo_lib.ApiClient)

	req := clo_ip.AddressDetailRequest{AddressID: fipId}
	resp, err := req.Do(ctx, cli)
	if err != nil {
		return diag.FromErr(err)
	}

	if e := d.Set("id", resp.Result.ID); e != nil {
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
	if e := d.Set("bandwidth", resp.Result.Bandwidth); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("is_primary", resp.Result.IsPrimary); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceIpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	if d.HasChange("ptr") {
		_, c := d.GetChange("ptr")
		if err := changeAddressPtr(ctx, d, c.(string), cli); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceIpDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressDeleteRequest{AddressID: d.Id()}
	if e := req.Do(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	if err := waitAddressDeleted(ctx, d.Id(), cli, d.Timeout(deletedIp)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// Waiters
func waitAddressState(ctx context.Context, id string, cli *clo_lib.ApiClient, pending []string, target []string, timeout time.Duration) (*clo_ip.AddressDetailResponse, error) {
	var resp *clo_ip.AddressDetailResponse
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_ip.AddressDetailRequest{AddressID: id}
			resp, err = req.Do(ctx, cli)
			return resp, resp.Result.Status, err
		},
		Pending:    pending,
		Target:     target,
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	err := resource.RetryContext(ctx, createStateConf.Timeout, func() *resource.RetryError {
		_, err := createStateConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
	return resp, err
}

func waitAddressDeleted(ctx context.Context, id string, cli *clo_lib.ApiClient, timeout time.Duration) error {
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_ip.AddressDetailRequest{AddressID: id}
			resp, err := req.Do(ctx, cli)

			apiError := clo_tools.DefaultError{}
			resState := resp.Result.Status
			if errors.As(err, &apiError) && apiError.Code == 404 {
				resState = deletedIp
				err = nil
			}
			return resp.Result, resState, err
		},
		Pending:    []string{processingIp},
		Target:     []string{deletedIp},
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, createStateConf.Timeout, func() *resource.RetryError {
		_, err := createStateConf.WaitForStateContext(ctx)
		if err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}

// Api actions
func changeAddressPtr(ctx context.Context, d *schema.ResourceData, prt string, cli *clo_lib.ApiClient) error {
	req := clo_ip.AddressPtrChangeRequest{
		AddressID: d.Id(),
		Body:      clo_ip.AddressPtrChangeBody{Value: prt},
	}
	return req.Do(ctx, cli)
}
