package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasNodes() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of nodes in a dbaas cluster",
		ReadContext: dataSourceDbaasNodesRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "ID of the dbaas cluster that owns the nodes",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id":         {Type: schema.TypeString, Computed: true},
						"name":       {Type: schema.TypeString, Computed: true},
						"cluster_id": {Type: schema.TypeString, Computed: true},
						"role":       {Type: schema.TypeString, Computed: true},
						"status":     {Type: schema.TypeString, Computed: true},
						"private_ip": {Type: schema.TypeString, Computed: true},
						"created_in": {Type: schema.TypeString, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceDbaasNodesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	nodes, err := cli.ListNodes(ctx, d.Get("cluster_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(nodes))
	for _, n := range nodes {
		res = append(res, map[string]interface{}{
			"id":         n.ID,
			"name":       n.Name,
			"cluster_id": n.ClusterID,
			"role":       n.Role,
			"status":     n.Status,
			"private_ip": n.PrivateIP,
			"created_in": n.CreatedIn,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
