package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceImages() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of the OS images",
		ReadContext: dataSourceImagesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns images",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Type:        schema.TypeList,
				Description: "The object that holds the results",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the image",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Name of the image",
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
				},
			},
		},
	}
}

func dataSourceImagesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	images, err := cli.ListImages(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(images))
	for _, im := range images {
		res = append(res, map[string]interface{}{
			"id":              im.ID,
			"name":            im.Name,
			"os_family":       im.OSFamily,
			"os_version":      im.OSVersion,
			"os_distribution": im.OSDistribution,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
