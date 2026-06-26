package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceS3User() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the data of an user of the object storage with a provided ID",
		ReadContext: dataSourceS3UserRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "ID of the user",
				Type:        schema.TypeString, Required: true},
			"name": {
				Description: "Human-readable name of the user",
				Type:        schema.TypeString, Computed: true},
			"tenant": {
				Description: "Name of the user's tenant. " +
					"Name of the user's project uses by default",
				Type: schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"canonical_name": {
				Description: "Canonical name of the user. The storage uses this name. " +
					"Should be unique in scope of the tenant",
				Type: schema.TypeString, Computed: true},
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
	}
}

func dataSourceS3UserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	user, err := cli.GetS3User(ctx, d.Get("user_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	fields := map[string]interface{}{
		"name":           user.Name,
		"status":         user.Status,
		"tenant":         user.Tenant,
		"max_buckets":    user.MaxBuckets,
		"canonical_name": user.CanonicalName,
		"quotas":         flattenS3Quota(user.Quotas),
	}
	for k, v := range fields {
		if e := d.Set(k, v); e != nil {
			return diag.FromErr(e)
		}
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}

func flattenS3Quota(quotas []cloapi.S3Quota) []interface{} {
	out := make([]interface{}, 0, len(quotas))
	for _, q := range quotas {
		out = append(out, map[string]interface{}{
			"type":        q.Type,
			"max_size":    q.MaxSize,
			"max_objects": q.MaxObjects,
		})
	}
	return out
}
