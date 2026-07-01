package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVolumes() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of volumes",
		ReadContext: dataSourceVolumesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns volumes",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the volume",
							Type:        schema.TypeString, Computed: true},
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
						"attached_to_instance_id": {
							Description: "ID of an instance the volume attached",
							Type:        schema.TypeString, Computed: true},
						"bootable": {
							Description: "Is the volume bootable",
							Type:        schema.TypeBool, Computed: true},
						"undetachable": {
							Description: "Can the volume be detached from the instance",
							Type:        schema.TypeBool, Computed: true},
						"size": {
							Description: "Size of the volume in Gb",
							Type:        schema.TypeInt, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceVolumesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	volumes, err := cli.ListVolumes(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("result", flattenVolumesResults(volumes)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenVolumesResults(pr []cloapi.Volume) []interface{} {
	res := make([]interface{}, 0, len(pr))
	for _, p := range pr {
		ri := map[string]interface{}{
			"id":           p.ID,
			"name":         p.Name,
			"status":       p.Status,
			"bootable":     p.Bootable,
			"undetachable": p.Undetachable,
			"created_in":   p.CreatedIn,
			"size":         p.Size,
		}
		if p.Attachment != nil {
			ri["device"] = p.Attachment.Device
			ri["attached_to_instance_id"] = p.Attachment.ID
		}
		res = append(res, ri)
	}
	return res
}
