package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSnapshots() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of server snapshots in the project",
		ReadContext: dataSourceSnapshotsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns the snapshots",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":            {Type: schema.TypeString, Computed: true},
						"name":          {Type: schema.TypeString, Computed: true},
						"status":        {Type: schema.TypeString, Computed: true},
						"size":          {Type: schema.TypeInt, Computed: true},
						"parent_server": {Type: schema.TypeString, Computed: true},
						"child_servers": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"created_in": {Type: schema.TypeString, Computed: true},
						"deleted_in": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceSnapshotsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	snapshots, err := cli.ListSnapshots(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(snapshots))
	for _, s := range snapshots {
		res = append(res, map[string]interface{}{
			"id":            s.ID,
			"name":          s.Name,
			"status":        s.Status,
			"size":          s.Size,
			"parent_server": s.ParentServer,
			"child_servers": s.ChildServers,
			"created_in":    s.CreatedIn,
			"deleted_in":    s.DeletedIn,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
