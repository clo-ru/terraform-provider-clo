package clo

import (
	"context"
	"errors"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_tools "github.com/clo-ru/cloapi-go-client/v2/clo/request_tools"
	clo_disks "github.com/clo-ru/cloapi-go-client/v2/services/disks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"strings"
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
	resp, err := createVolume(ctx, cli, d)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = waitVolumeState(ctx, resp.Result.ID, cli, []string{creatingVolume}, []string{activeVolume}, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Result.ID)
	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	if d.HasChange("size") {
		_, c := d.GetChange("size")
		if err := resizeVolume(ctx, cli, d, c.(int)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeDeleteRequest{VolumeID: d.Id()}
	if e := req.Do(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	err := waitVolumeDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete))
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
	resp, e := req.Do(ctx, cli)
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

// Waiters
func waitVolumeState(ctx context.Context, id string, cli *clo_lib.ApiClient, pending []string, target []string, timeout time.Duration) (*clo_disks.VolumeDetailResponse, error) {
	var resp *clo_disks.VolumeDetailResponse
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_disks.VolumeDetailRequest{VolumeID: id}
			resp, err = req.Do(ctx, cli)
			resState := strings.ToUpper(resp.Result.Status)
			return resp, resState, err
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

func waitVolumeDeleted(ctx context.Context, id string, cli *clo_lib.ApiClient, timeout time.Duration) error {
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_disks.VolumeDetailRequest{VolumeID: id}
			resp, err := req.Do(ctx, cli)

			apiError := clo_tools.DefaultError{}
			resState := resp.Result.Status
			if errors.As(err, &apiError) && apiError.Code == 404 {
				resState = deletedVolume
				err = nil
			}
			resState = strings.ToUpper(resState)
			return resp.Result, resState, err
		},
		Pending:    []string{deletingVolume},
		Target:     []string{deletedVolume},
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
func createVolume(ctx context.Context, cli *clo_lib.ApiClient, d *schema.ResourceData) (*clo_lib.ResponseCreated, error) {
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
	return req.Do(ctx, cli)
}

func resizeVolume(ctx context.Context, cli *clo_lib.ApiClient, d *schema.ResourceData, size int) error {
	req := clo_disks.VolumeResizeRequest{
		VolumeID: d.Id(),
		Body:     clo_disks.VolumeResizeBody{NewSize: size},
	}
	if err := req.Do(ctx, cli); err != nil {
		return err
	}
	_, err := waitVolumeState(ctx, d.Id(), cli, []string{resizingVolume}, []string{activeVolume, attachedVolume}, d.Timeout(schema.TimeoutCreate))
	return err
}
