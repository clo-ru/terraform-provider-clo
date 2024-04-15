package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_project "github.com/clo-ru/cloapi-go-client/v2/services/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
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
			"results": {
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
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_project.ImageListRequest{
		ProjectID: d.Get("project_id").(string),
	}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	var res []interface{}
	lr := len(resp.Result)
	if lr > 0 {
		res = make([]interface{}, lr)
		for i, r := range resp.Result {
			m := map[string]interface{}{"id": r.ID, "name": r.Name}
			if osSystem := r.OperationSystem; osSystem != nil {
				m["os_family"] = r.OperationSystem.OsFamily
				m["os_version"] = r.OperationSystem.Version
				m["os_distribution"] = r.OperationSystem.Distribution
			}
			res[i] = m
		}
	}
	if e := d.Set("results", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}
