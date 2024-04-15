package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_disks "github.com/clo-ru/cloapi-go-client/v2/services/disks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		},
	}
}

func resourceVolumeAttachCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Get("volume_id").(string)
	sid := d.Get("instance_id").(string)
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeAttachRequest{
		VolumeID: vid,
		Body:     clo_disks.VolumeAttachBody{ServerID: sid},
	}
	e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}

	resp, err := waitVolumeState(ctx, vid, cli, []string{attachingVolume}, []string{attachedVolume}, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	if e := d.Set("device", resp.Result.Attachment.Device); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(vid)
	return nil
}

func resourceVolumeAttachRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vid := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeDetailRequest{VolumeID: vid}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}

	device := ""
	if resp.Result.Attachment != nil {
		device = resp.Result.Attachment.ID
	}
	if e := d.Set("device", device); e != nil {
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
	if e := req.Do(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	_, err := waitVolumeState(ctx, vid, cli, []string{detachingVolume}, []string{activeVolume}, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
