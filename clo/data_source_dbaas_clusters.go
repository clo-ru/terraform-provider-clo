package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasClusters() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of dbaas clusters in the project",
		ReadContext: dataSourceDbaasClustersRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns dbaas clusters",
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
						"status":            {Type: schema.TypeString, Computed: true},
						"switch_status":     {Type: schema.TypeString, Computed: true},
						"datastore_id":      {Type: schema.TypeString, Computed: true},
						"datastore_name":    {Type: schema.TypeString, Computed: true},
						"datastore_version": {Type: schema.TypeString, Computed: true},
						"storage_size":      {Type: schema.TypeInt, Computed: true},
						"storage_used_kb":   {Type: schema.TypeInt, Computed: true},
						"system_disk_size":  {Type: schema.TypeInt, Computed: true},
						"nodes_count":       {Type: schema.TypeInt, Computed: true},
						"databases_count":   {Type: schema.TypeInt, Computed: true},
						"external_address":  {Type: schema.TypeString, Computed: true},
						"internal_address":  {Type: schema.TypeString, Computed: true},
						"backup_hour":       {Type: schema.TypeInt, Computed: true},
						"created_in":        {Type: schema.TypeString, Computed: true},
						"flavor": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"vcpus": {Type: schema.TypeInt, Computed: true},
								"ram":   {Type: schema.TypeInt, Computed: true},
								"disk":  {Type: schema.TypeInt, Computed: true},
							}},
						},
					},
				},
			},
		},
	}
}

func dataSourceDbaasClustersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	clusters, err := cli.ListClusters(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(clusters))
	for i := range clusters {
		c := clusters[i]
		res = append(res, map[string]interface{}{
			"id":                c.ID,
			"name":              c.Name,
			"status":            c.Status,
			"switch_status":     c.SwitchStatus,
			"datastore_id":      c.DatastoreID,
			"datastore_name":    c.DatastoreName,
			"datastore_version": c.DatastoreVersion,
			"storage_size":      c.StorageSize,
			"storage_used_kb":   c.StorageUsedKB,
			"system_disk_size":  c.SystemDiskSize,
			"nodes_count":       c.NodesCount,
			"databases_count":   c.DatabasesCount,
			"external_address":  c.ExternalAddress,
			"internal_address":  c.InternalAddress,
			"backup_hour":       c.BackupHour,
			"created_in":        c.CreatedIn,
			"flavor":            flattenClusterFlavor(&c),
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
