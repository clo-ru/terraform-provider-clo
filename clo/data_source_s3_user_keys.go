package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	cli := m.(*providerMeta).v3
	// The v3 API returns only the access key on read; the secret key is available
	// only at key generation, so secret_key stays empty here.
	accessKey, err := cli.GetS3UserAccessKey(ctx, d.Get("user_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("access_key", accessKey); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
