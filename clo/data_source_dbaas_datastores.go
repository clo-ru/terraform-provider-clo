package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasDatastores() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of dbaas datastores (database engines + versions) available in the project",
		ReadContext: dataSourceDbaasDatastoresRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":      {Type: schema.TypeString, Computed: true},
						"name":    {Type: schema.TypeString, Computed: true},
						"version": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceDbaasDatastoresRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	datastores, err := cli.ListDatastores(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(datastores))
	for _, ds := range datastores {
		res = append(res, map[string]interface{}{
			"id":      ds.ID,
			"name":    ds.Name,
			"version": ds.Version,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
