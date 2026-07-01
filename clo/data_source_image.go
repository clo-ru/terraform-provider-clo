package clo

import (
	"context"

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
	cli := m.(*providerMeta).v3
	images, err := cli.ListImages(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	for _, im := range images {
		if im.Name != name {
			continue
		}
		fields := map[string]interface{}{
			"image_id":        im.ID,
			"os_distribution": im.OSDistribution,
			"os_version":      im.OSVersion,
			"os_family":       im.OSFamily,
		}
		for k, v := range fields {
			if e := d.Set(k, v); e != nil {
				return diag.FromErr(e)
			}
		}
		d.SetId(im.ID)
		break
	}
	return nil
}
