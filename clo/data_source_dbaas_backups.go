package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasBackups() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of dbaas backups in the project",
		ReadContext: dataSourceDbaasBackupsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns the backups",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":                {Type: schema.TypeString, Computed: true},
						"name":              {Type: schema.TypeString, Computed: true},
						"cluster_id":        {Type: schema.TypeString, Computed: true},
						"project":           {Type: schema.TypeString, Computed: true},
						"status":            {Type: schema.TypeString, Computed: true},
						"type":              {Type: schema.TypeString, Computed: true},
						"size":              {Type: schema.TypeInt, Computed: true},
						"data_size":         {Type: schema.TypeInt, Computed: true},
						"parent":            {Type: schema.TypeString, Computed: true},
						"datastore_name":    {Type: schema.TypeString, Computed: true},
						"datastore_version": {Type: schema.TypeString, Computed: true},
						"created_in":        {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceDbaasBackupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	backups, err := cli.ListBackups(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(backups))
	for _, b := range backups {
		res = append(res, map[string]interface{}{
			"id":                b.ID,
			"name":              b.Name,
			"cluster_id":        b.ClusterID,
			"project":           b.Project,
			"status":            b.Status,
			"type":              b.Type,
			"size":              b.Size,
			"data_size":         b.DataSize,
			"parent":            b.Parent,
			"datastore_name":    b.DatastoreName,
			"datastore_version": b.DatastoreVersion,
			"created_in":        b.CreatedIn,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
