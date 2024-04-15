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

func dataSourceS3Keys() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the data of S3 user's keys",
		ReadContext: dataSourceS3KeysRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceS3KeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	cli := m.(*clo_lib.ApiClient)
	req := clo_storage.S3KeysGetRequest{UserID: d.Get("user_id").(string)}
	resp, e := req.Do(ctx, cli)

	if e != nil {
		return diag.FromErr(e)
	}
	if len(resp.Result) > 0 {
		key := resp.Result[0]
		if e := d.Set("access_key", key.AccessKey); e != nil {
			return diag.FromErr(e)
		}
		if e := d.Set("secret_key", key.SecretKey); e != nil {
			return diag.FromErr(e)
		}
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}
