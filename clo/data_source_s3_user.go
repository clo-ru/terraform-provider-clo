package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_storage "github.com/clo-ru/cloapi-go-client/services/storage"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"time"
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
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_storage.S3UserDetailRequest{
		UserID: d.Get("user_id").(string),
	}
	resp, e := req.Make(ctx, cli)
	if resp.Code == 404 {
		e = fmt.Errorf("NotFound returned")
	}
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("name", resp.Result.Name); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("tenant", resp.Result.Tenant); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("max_buckets", resp.Result.MaxBuckets); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("canonical_name", resp.Result.CanonicalName); e != nil {
		return diag.FromErr(e)
	}
	var quotaData []interface{}
	ld := len(resp.Result.Quotas)
	if ld > 0 {
		quotaData = make([]interface{}, ld)
		for j, q := range resp.Result.Quotas {
			quotaData[j] = map[string]interface{}{
				"type":        q.Type,
				"max_size":    q.MaxSize,
				"max_objects": q.MaxObjects,
			}
		}
	}
	if e := d.Set("quotas", quotaData); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}
