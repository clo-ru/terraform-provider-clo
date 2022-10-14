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
	req := clo_storage.S3KeysGetRequest{
		UserID: d.Get("user_id").(string),
	}
	resp, e := req.Make(ctx, cli)
	if resp.Code == 404 {
		e = fmt.Errorf("NotFound returned")
	}
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("access_key", resp.Result.AccessKey); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("secret_key", resp.Result.SecretKey); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}
