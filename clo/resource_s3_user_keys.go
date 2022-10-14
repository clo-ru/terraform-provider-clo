package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_storage "github.com/clo-ru/cloapi-go-client/services/storage"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceS3UserKeys() *schema.Resource {
	return &schema.Resource{
		Description:   "Create key's pairs for the user",
		ReadContext:   resourceS3UserKeysRead,
		CreateContext: resourceS3UserKeysCreate,
		DeleteContext: resourceS3UserKeysDelete,
		Timeouts: &schema.ResourceTimeout{
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "ID of the user for whom the keys will be generated",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
			},
			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
				ForceNew:  true,
			},
		},
	}
}

func resourceS3UserKeysCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	uId := d.Get("user_id").(string)
	cli := m.(*clo_lib.ApiClient)
	req := clo_storage.S3KeysResetRequest{
		UserID: uId,
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	d.SetId(uId)
	return setKeysResult(d, resp.Result)
}

func resourceS3UserKeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	uId := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_storage.S3KeysGetRequest{
		UserID: uId,
	}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	return setKeysResult(d, resp.Result)
}

func setKeysResult(d *schema.ResourceData, resp clo_storage.S3KeysResponse) diag.Diagnostics {
	if e := d.Set("access_key", resp.AccessKey); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("secret_key", resp.SecretKey); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceS3UserKeysDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
