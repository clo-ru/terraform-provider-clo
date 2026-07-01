package clo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKeypair() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches a keypair by name",
		ReadContext: dataSourceKeypairRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns the keypair",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The name of the desired keypair",
				Type:        schema.TypeString,
				Required:    true,
			},
			"keypair_id": {
				Description: "ID of the requested keypair",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"public_key": {
				Description: "Public key of the keypair",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_in": {
				Description: "Timestamp the keypair was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceKeypairRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	keypairs, err := cli.ListKeypairs(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	for _, kp := range keypairs {
		if kp.Name != name {
			continue
		}
		fields := map[string]interface{}{
			"keypair_id": kp.ID,
			"public_key": kp.PublicKey,
			"created_in": kp.CreatedIn,
		}
		for k, v := range fields {
			if e := d.Set(k, v); e != nil {
				return diag.FromErr(e)
			}
		}
		d.SetId(kp.ID)
		break
	}
	return nil
}
