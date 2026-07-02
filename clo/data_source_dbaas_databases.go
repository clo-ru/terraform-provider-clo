package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasDatabases() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of databases in a dbaas cluster",
		ReadContext: dataSourceDbaasDatabasesRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "ID of the dbaas cluster that owns the databases",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":             {Type: schema.TypeString, Computed: true},
						"name":           {Type: schema.TypeString, Computed: true},
						"cluster_id":     {Type: schema.TypeString, Computed: true},
						"project":        {Type: schema.TypeString, Computed: true},
						"admin_username": {Type: schema.TypeString, Computed: true},
						"status":         {Type: schema.TypeString, Computed: true},
						"backup_enabled": {Type: schema.TypeBool, Computed: true},
						"created_in":     {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceDbaasDatabasesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	databases, err := cli.ListDatabasesByCluster(ctx, d.Get("cluster_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(databases))
	for _, db := range databases {
		res = append(res, map[string]interface{}{
			"id":             db.ID,
			"name":           db.Name,
			"cluster_id":     db.ClusterID,
			"project":        db.Project,
			"admin_username": db.AdminUsername,
			"status":         db.Status,
			"backup_enabled": db.BackupEnabled,
			"created_in":     db.CreatedIn,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
