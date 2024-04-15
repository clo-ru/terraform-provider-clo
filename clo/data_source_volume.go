package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_disks "github.com/clo-ru/cloapi-go-client/v2/services/disks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceVolume() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the data of an volume with a provided ID",
		ReadContext: dataSourceVolumeRead,
		Schema: map[string]*schema.Schema{
			"volume_id": {
				Description: "ID of the volume",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Human-readable name of the volume",
				Type:        schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"device": {
				Description: "Represents the filesystem's device name, for example: `/dev/vdb`",
				Type:        schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the volume was created",
				Type:        schema.TypeString, Computed: true},
			"description": {Type: schema.TypeString, Computed: true},
			"attached_to_instance_id": {
				Description: "ID of an instance the volume attached",
				Type:        schema.TypeString, Computed: true},
			"size": {
				Description: "Size of the volume in Gb",
				Type:        schema.TypeInt, Computed: true},
			"bootable": {
				Description: "Is the volume bootable",
				Type:        schema.TypeBool, Computed: true},
			"undetachable": {
				Description: "Can the volume be detached from the instance",
				Type:        schema.TypeBool, Computed: true},
		},
	}
}

func dataSourceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeDetailRequest{
		VolumeID: d.Get("volume_id").(string),
	}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("name", resp.Result.Name); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", resp.Result.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("description", resp.Result.Description); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("size", resp.Result.Size); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("bootable", resp.Result.Bootable); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("undetachable", resp.Result.Undetachable); e != nil {
		return diag.FromErr(e)
	}
	if resp.Result.Attachment != nil {
		if e := d.Set("device", resp.Result.Attachment.Device); e != nil {
			return diag.FromErr(e)
		}
		if e := d.Set("attached_to_instance_id", resp.Result.Attachment.ID); e != nil {
			return diag.FromErr(e)
		}
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
