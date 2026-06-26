package clo

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	cli := m.(*providerMeta).v3
	keys, err := cli.GenS3UserKeys(ctx, uId)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(uId)
	if e := d.Set("access_key", keys.AccessKey); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("secret_key", keys.SecretKey); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceS3UserKeysRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	accessKey, err := cli.GetS3UserAccessKey(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	// Only the access key is returned on read; the secret key (set at create) is left as-is.
	if e := d.Set("access_key", accessKey); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

func resourceS3UserKeysDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}
