package clo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasBackupDownload() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches a presigned download URL for a dbaas backup. The URL is ephemeral, so it is refreshed on every read.",
		ReadContext: dataSourceDbaasBackupDownloadRead,
		Schema: map[string]*schema.Schema{
			"backup_id": {
				Description: "ID of the backup to download",
				Type:        schema.TypeString,
				Required:    true,
			},
			"url": {
				Description: "Presigned URL the backup can be downloaded from",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceDbaasBackupDownloadRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id := d.Get("backup_id").(string)
	url, err := cli.DownloadBackup(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("url", url); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(id)
	return nil
}
