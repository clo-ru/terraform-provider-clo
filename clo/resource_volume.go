package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_disks "github.com/clo-ru/cloapi-go-client/services/disks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

const (
	creatingVolume  = "CREATING"
	activeVolume    = "AVAILABLE"
	resizingVolume  = "RESIZING"
	attachingVolume = "ATTACHING"
	attachedVolume  = "IN-USE"
	detachingVolume = "DETACHING"
	deletingVolume  = "DELETING"
	deletedVolume   = "DELETED"
)

func resourceVolume() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a new volume in the project",
		ReadContext:   resourceVolumeRead,
		CreateContext: resourceVolumeCreate,
		UpdateContext: resourceVolumeUpdate,
		DeleteContext: resourceVolumeDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("size", func(ctx context.Context, oldValue, newValue, meta interface{}) error {
				if newValue.(int) < oldValue.(int) {
					return fmt.Errorf("size could be increased only")
				}
				return nil
			})),
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the volume should be created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Human-readable name of the new volume",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"size": {
				Description: "Size of the new volume in Gb",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
					if sz := i.(int); sz < 10 {
						errs = append(errs, fmt.Errorf("size should be at least 10Gb"))
					}
					return
				},
			},
			"id": {
				Description: "ID of the new volume",
				Type:        schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the volume was created",
				Type:        schema.TypeString, Computed: true},
		},
	}
}

func resourceVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeCreateRequest{
		ProjectID: d.Get("project_id").(string),
		Body: clo_disks.VolumeCreateBody{
			Size: d.Get("size").(int),
		},
	}
	if n, ok := d.GetOk("name"); ok {
		req.Body.Name = n.(string)
	} else {
		req.Body.Autorename = true
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
			req := clo_disks.VolumeDetailRequest{VolumeID: resp.Result.ID}
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
		Target:     []string{activeVolume},
		Pending:    []string{creatingVolume},
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Result.ID)
	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	stateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_disks.VolumeDetailRequest{VolumeID: d.Id()}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			} else {
				return resp, resp.Result.Status, nil
			}
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Minute,
	}
	if d.HasChange("size") {
		_, c := d.GetChange("size")
		req := clo_disks.VolumeResizeRequest{
			VolumeID: d.Id(),
			Body:     clo_disks.VolumeResizeBody{NewSize: c.(int)},
		}
		stateConf.Pending = []string{resizingVolume}
		stateConf.Target = []string{activeVolume}
		if e := req.Make(ctx, cli); e != nil {
			return diag.FromErr(e)
		}
		_, err := stateConf.WaitForStateContext(ctx)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeDeleteRequest{VolumeID: d.Id()}
	if e := req.Make(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_disks.VolumeDetailRequest{VolumeID: d.Id()}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			}
			if resp.Code == 404 {
				return resp.Result, deletedVolume, nil
			}
			return resp.Result, resp.Result.Status, nil
		},
		Target:     []string{deletedVolume},
		Pending:    []string{deletingVolume},
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

func resourceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeDetailRequest{
		VolumeID: vid,
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
	if e := d.Set("created_in", resp.Result.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	return nil
}
