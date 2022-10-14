package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_disks "github.com/clo-ru/cloapi-go-client/services/disks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceVolumeAttach() *schema.Resource {
	return &schema.Resource{
		Description:   "Attach the volume to an instance",
		ReadContext:   resourceVolumeAttachRead,
		CreateContext: resourceVolumeAttachCreate,
		DeleteContext: resourceVolumeDetach,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"volume_id": {
				Description: "ID of the volume which should be attached",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"instance_id": {
				Description: "ID of the instance to which the volume will be attached",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"device": {
				Description: "Represents the filesystem's device name, for example: `/dev/vdb`",
				Type:        schema.TypeString, Computed: true},
			"mount_point_base": {Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceVolumeAttachCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Get("volume_id").(string)
	sid := d.Get("instance_id").(string)
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeAttachRequest{
		VolumeID: vid,
		Body: clo_disks.VolumeAttachBody{
			ServerID: sid,
		},
	}
	if n, ok := d.GetOk("mount_point_base"); ok {
		req.Body.MountPath = n.(string)
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_disks.VolumeDetailRequest{VolumeID: vid}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			}
			return resp, resp.Result.Status, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Target:     []string{attachedVolume},
		Pending:    []string{attachingVolume},
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("device", resp.Result.Device); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("mount_point_base", resp.Result.Mountpoint); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(vid)
	return nil
}

func resourceVolumeAttachRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeDetailRequest{
		VolumeID: vid,
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("device", vid); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("mount_point_base", resp.Result.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceVolumeDetach(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeDetachRequest{
		VolumeID: vid,
		Body: clo_disks.VolumeDetachBody{
			Force: true,
		},
	}
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
			return resp.Result, resp.Result.Status, nil
		},
		Target:     []string{activeVolume},
		Pending:    []string{detachingVolume},
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
