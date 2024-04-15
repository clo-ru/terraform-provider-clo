package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_project "github.com/clo-ru/cloapi-go-client/v2/services/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceImage() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of the OS images",
		ReadContext: dataSourceImageRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns images",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The name of the desired image",
				Type:        schema.TypeString,
				Required:    true,
			},
			"image_id": {
				Description: "ID of the requested image",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"os_distribution": {
				Description: "The distributed OS name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"os_family": {
				Description: "The family of the distributed OS",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"os_version": {
				Description: "The version of the distributed OS",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_project.ImageListRequest{
		ProjectID: d.Get("project_id").(string),
	}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	n := d.Get("name").(string)
	for _, r := range resp.Result {
		if r.Name == n {
			if e := d.Set("image_id", r.ID); e != nil {
				return diag.FromErr(e)
			}
			if osSystem := r.OperationSystem; osSystem != nil {
				if e := d.Set("os_distribution", osSystem.Distribution); e != nil {
					return diag.FromErr(e)
				}
				if e := d.Set("os_version", osSystem.Version); e != nil {
					return diag.FromErr(e)
				}
				if e := d.Set("os_family", osSystem.OsFamily); e != nil {
					return diag.FromErr(e)
				}
			}
			break
		}
		d.SetId(r.ID)
	}
	return diags
}
