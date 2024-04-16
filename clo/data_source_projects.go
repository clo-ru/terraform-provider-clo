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

func dataSourceProjects() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of the projects",
		ReadContext: dataSourceProjectsRead,
		Schema: map[string]*schema.Schema{
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the project",
							Type:        schema.TypeString, Computed: true},
						"name": {
							Description: "Name of the project",
							Type:        schema.TypeString, Computed: true},
						"status": {Type: schema.TypeString, Computed: true},
						"created_in": {
							Description: "Timestamp the project was created",
							Type:        schema.TypeString, Computed: true},
						"stopping_reason": {
							Description: "A reason the project was stopped",
							Type:        schema.TypeString, Computed: true},
						"has_abuse": {
							Description: "Is the project has an abuse issues",
							Type:        schema.TypeBool, Computed: true},
					},
				},
			},
		},
	}
}

func dataSourceProjectsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_project.ProjectListRequest{}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}

	if e := d.Set("results", flattenProjectResults(resp.Result)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenProjectResults(pr []clo_project.Project) []interface{} {
	lpr := len(pr)
	if lpr > 0 {
		res := make([]interface{}, lpr, lpr)
		for i, p := range pr {
			ri := make(map[string]interface{})
			ri["id"] = p.ID
			ri["name"] = p.Name
			ri["status"] = p.Status
			ri["has_abuse"] = p.HasAbuse
			ri["created_in"] = p.CreatedIn
			ri["stopping_reason"] = p.StoppingReason
			res[i] = ri
		}
		return res
	}
	return make([]interface{}, 0)
}
