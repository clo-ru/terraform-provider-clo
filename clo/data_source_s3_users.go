package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	clo_storage "github.com/clo-ru/cloapi-go-client/v2/services/storage"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
)

func dataSourceS3Users() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of users of the object storage",
		ReadContext: dataSourceS3UsersRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns users",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the user",
							Type:        schema.TypeString, Computed: true},
						"name": {
							Description: "Human-readable name of the user",
							Type:        schema.TypeString, Computed: true},
						"canonical_name": {
							Description: "Canonical name of the user. Storage uses this name. " +
								"Should be unique in scope of the tenant",
							Type: schema.TypeString, Computed: true},
						"tenant": {
							Description: "Name of the user's tenant. " +
								"Name of the user's project uses by default",
							Type: schema.TypeString, Computed: true},
						"status": {Type: schema.TypeString, Computed: true},
						"max_buckets": {
							Description: "How many buckets the user could create",
							Type:        schema.TypeInt, Computed: true},
						"quotas": {
							Description: "The object represents a user's quota",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"max_objects": {
									Description: "How many objects the user can create",
									Type:        schema.TypeInt, Computed: true},
								"max_size": {
									Description: "Total size of the objects the user can store",
									Type:        schema.TypeInt, Computed: true},
								"type": {
									Description: "Could be USER or BUCKET",
									Type:        schema.TypeString, Computed: true},
							}}},
					},
				},
			},
		},
	}
}

func dataSourceS3UsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_storage.S3UserListRequest{
		ProjectID: d.Get("project_id").(string),
	}
	resp, e := req.Do(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("result", flattenS3UsersResults(resp.Result)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func flattenS3UsersResults(pr []clo_storage.S3User) []interface{} {
	lpr := len(pr)
	if lpr > 0 {
		res := make([]interface{}, lpr, lpr)
		for i, p := range pr {
			ri := make(map[string]interface{})
			ri["id"] = p.ID
			ri["name"] = p.Name
			ri["status"] = p.Status
			ri["tenant"] = p.Tenant
			ri["max_buckets"] = p.MaxBuckets
			ri["canonical_name"] = p.CanonicalName
			ri["quotas"] = formatS3Quota(p.Quotas)
			res[i] = ri
		}
		return res
	}
	return make([]interface{}, 0)
}
