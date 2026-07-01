package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	cli := m.(*providerMeta).v3
	users, err := cli.ListS3Users(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("result", flattenS3UsersResults(users)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenS3UsersResults(pr []cloapi.S3User) []interface{} {
	res := make([]interface{}, 0, len(pr))
	for _, p := range pr {
		res = append(res, map[string]interface{}{
			"id":             p.ID,
			"name":           p.Name,
			"status":         p.Status,
			"tenant":         p.Tenant,
			"max_buckets":    p.MaxBuckets,
			"canonical_name": p.CanonicalName,
			"quotas":         flattenS3Quota(p.Quotas),
		})
	}
	return res
}
