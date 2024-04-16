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
							Description: "ID of the user",
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
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_disks.VolumeListRequest{
		ProjectID: d.Get("project_id").(string),
	}
	resp, e := req.Do(ctx, cli)

	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("result", flattenVolumesResults(resp.Result)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenVolumesResults(pr []clo_disks.Volume) []interface{} {
	lpr := len(pr)
	if lpr > 0 {
		res := make([]interface{}, lpr, lpr)
		for i, p := range pr {
			ri := make(map[string]interface{})
			ri["id"] = p.ID
			ri["name"] = p.Name

			ri["status"] = p.Status
			ri["bootable"] = p.Bootable
			ri["undetachable"] = p.Undetachable
			ri["created_in"] = p.CreatedIn
			ri["size"] = p.Size

			if p.Attachment != nil {
				ri["device"] = p.Attachment.Device
				ri["attached_to_instance_id"] = p.Attachment.ID
			}

			res[i] = ri
		}
		return res
	}
	return make([]interface{}, 0)
}
