package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_ip "github.com/clo-ru/cloapi-go-client/services/ip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"primary": {
				Description: "Should the address be used as primary",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"ptr": {
				Description: "PTR of the attached address",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"id": {
				Description: "ID of the created address",
				Type:        schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"address": {
				Description: "String representation of the address",
				Type:        schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the address was created",
				Type:        schema.TypeString, Computed: true},
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
	resp, e := req.Make(ctx, cli)
	if resp.Code == 404 {
		e = fmt.Errorf("NotFound returned")
	}
	if e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_ip.AddressDetailRequest{
				AddressID: resp.Result.ID,
			}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			} else {
				return resp, resp.Result.Status, nil
			}
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Target:     []string{detachedIp},
		Pending:    []string{processingIp},
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if n, ok := d.GetOk("ptr"); ok {
		req := clo_ip.AddressPtrChangeRequest{
			AddressID: d.Id(),
			Body:      clo_ip.AddressPtrChangeBody{Value: n.(string)},
		}
		if e := changePtr(req, cli); e != nil {
			return diag.FromErr(e)
		}
	}
	d.SetId(resp.Result.ID)
	return resourceIpRead(ctx, d, m)
}

func resourceIpRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	fipId := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressDetailRequest{
		AddressID: fipId,
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
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
	return nil
}

func resourceIpUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	if d.HasChange("ptr") {
		_, c := d.GetChange("ptr")
		req := clo_ip.AddressPtrChangeRequest{
			AddressID: d.Id(),
			Body:      clo_ip.AddressPtrChangeBody{Value: c.(string)},
		}
		if e := changePtr(req, cli); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func changePtr(req clo_ip.AddressPtrChangeRequest, cli *clo_lib.ApiClient) error {
	if e := req.Make(context.Background(), cli); e != nil {
		return e
	}
	return nil
}

func resourceIpDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressDeleteRequest{AddressID: d.Id()}
	if e := req.Make(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_ip.AddressDetailRequest{AddressID: d.Id()}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			}
			if resp.Code == 404 {
				return resp.Result, deletedIp, nil
			}
			return resp.Result, resp.Result.Status, nil
		},
		Target:     []string{deletedIp},
		Pending:    []string{processingIp},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
