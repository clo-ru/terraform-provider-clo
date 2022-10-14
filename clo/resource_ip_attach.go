package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_ip "github.com/clo-ru/cloapi-go-client/services/ip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceIpAttach() *schema.Resource {
	return &schema.Resource{
		Description:   "Attach an address to the entity, for example: a loadbalancer or a server",
		ReadContext:   resourceIpAttachRead,
		CreateContext: resourceIpAttachCreate,
		UpdateContext: resourceIpAttachUpdate,
		DeleteContext: resourceIpDetach,
		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"address_id": {
				Description: "ID of the attached address",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"entity_id": {
				Description: "ID of the entity the address will be attached to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"entity_name": {
				Description: "Name of the entity. Should be `loadbalancer` or `server`",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"is_primary": {
				Description: "Use the address as a primary address",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"status":  {Type: schema.TypeString, Computed: true},
			"address": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceIpAttachCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	adId := d.Get("address_id").(string)
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressAttachRequest{
		AddressID: adId,
		Body: clo_ip.AddressAttachBody{
			ID:     d.Get("entity_id").(string),
			Entity: d.Get("entity_name").(string),
		},
	}
	_, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	var detResp clo_ip.AddressDetailResponse
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			detReq := clo_ip.AddressDetailRequest{AddressID: adId}
			detResp, e = detReq.Make(ctx, cli)
			if e != nil {
				return detResp, "", e
			}
			stat := detResp.Result.Status
			if len(stat) == 0 {
				stat = processingIp
			}
			return detResp.Result, detResp.Result.Status, nil
		},
		Target:     []string{attachedIp},
		Pending:    []string{processingIp},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if _, ok := d.GetOk("is_primary"); ok {
		if e := makePrimary(ctx, adId, cli); e != nil {
			return diag.FromErr(e)
		}
	}
	if e := d.Set("status", detResp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("is_primary", detResp.Result.IsPrimary); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(adId)
	return resourceIpAttachRead(ctx, d, m)
}

func resourceIpAttachRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	adId := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressDetailRequest{
		AddressID: adId,
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("address", resp.Result.Address); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceIpDetach(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_ip.AddressDetachRequest{AddressID: d.Id()}
	if _, e := req.Make(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_ip.AddressDetailRequest{AddressID: d.Id()}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			}
			return resp.Result, resp.Result.Status, nil
		},
		Target:     []string{detachedIp},
		Pending:    []string{processingIp},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Minute,
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceIpAttachUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChange("is_primary") {
		if e := makePrimary(ctx, d.Id(), m.(*clo_lib.ApiClient)); e != nil {
			return diag.FromErr(e)
		}
	}
	return nil
}

func makePrimary(ctx context.Context, adId string, cli *clo_lib.ApiClient) error {
	req := clo_ip.AddressPrimaryChangeRequest{
		Request:   clo_lib.Request{},
		AddressID: adId,
	}
	_, e := req.Make(ctx, cli)
	if e != nil {
		return e
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_ip.AddressDetailRequest{AddressID: adId}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			}
			return resp.Result, resp.Result.Status, nil
		},
		Target:     []string{detachedIp},
		Pending:    []string{processingIp},
		Delay:      10 * time.Second,
		Timeout:    20 * time.Minute,
		MinTimeout: 1 * time.Minute,
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
