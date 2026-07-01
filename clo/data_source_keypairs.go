package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKeypairs() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of keypairs in the project",
		ReadContext: dataSourceKeypairsRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns keypairs",
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
							Description: "ID of the keypair",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "Name of the keypair",
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
				},
			},
		},
	}
}

func dataSourceKeypairsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	keypairs, err := cli.ListKeypairs(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(keypairs))
	for _, kp := range keypairs {
		res = append(res, map[string]interface{}{
			"id":         kp.ID,
			"name":       kp.Name,
			"public_key": kp.PublicKey,
			"created_in": kp.CreatedIn,
		})
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
