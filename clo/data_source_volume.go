package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	cli := m.(*providerMeta).v3
	vol, err := cli.GetVolume(ctx, d.Get("volume_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("name", vol.Name); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", vol.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", vol.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("description", vol.Description); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("size", vol.Size); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("bootable", vol.Bootable); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("undetachable", vol.Undetachable); e != nil {
		return diag.FromErr(e)
	}
	if vol.Attachment != nil {
		if e := d.Set("device", vol.Attachment.Device); e != nil {
			return diag.FromErr(e)
		}
		if e := d.Set("attached_to_instance_id", vol.Attachment.ID); e != nil {
			return diag.FromErr(e)
		}
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
