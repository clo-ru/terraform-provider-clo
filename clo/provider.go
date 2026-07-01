package clo

import (
	"context"
	"errors"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerMeta carries the SDK client passed to every resource and data source.
type providerMeta struct {
	v3 *cloapi.Client
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"auth_url": {
				Description: "URI for CLO API. May also be provided via CLO_API_AUTH_URL environment variable.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLO_API_AUTH_URL", nil),
			},
			"token": {
				Description: "JWT token. Should be issued in user area. " +
					"May also be provided via CLO_API_AUTH_TOKEN environment variable.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("CLO_API_AUTH_TOKEN", nil),
			},
		},
		ConfigureContextFunc: configureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"clo_compute_instance":          resourceInstance(),
			"clo_network_ip":                resourceIp(),
			"clo_network_ip_attach":         resourceIpAttach(),
			"clo_disks_volume":              resourceVolume(),
			"clo_disks_volume_attach":       resourceVolumeAttach(),
			"clo_storage_s3_user":           resourceS3User(),
			"clo_storage_s3_user_keys":      resourceS3UserKeys(),
			"clo_compute_keypair":           resourceKeypair(),
			"clo_network_vrouter":           resourceVrouter(),
			"clo_network_loadbalancer":      resourceLoadBalancer(),
			"clo_network_loadbalancer_rule": resourceLoadBalancerRule(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"clo_projects":             dataSourceProjects(),
			"clo_project_images":       dataSourceImages(),
			"clo_project_image":        dataSourceImage(),
			"clo_project_recipes":      dataSourceRecipes(),
			"clo_project_recipe":       dataSourceRecipe(),
			"clo_compute_keypair":      dataSourceKeypair(),
			"clo_compute_keypairs":     dataSourceKeypairs(),
			"clo_network_ip":           dataSourceIP(),
			"clo_network_ips":          dataSourceIPs(),
			"clo_disks_volume":         dataSourceVolume(),
			"clo_disks_volumes":        dataSourceVolumes(),
			"clo_compute_instance":     dataSourceInstance(),
			"clo_compute_instances":    dataSourceInstances(),
			"clo_storage_s3_user":      dataSourceS3User(),
			"clo_storage_s3_users":     dataSourceS3Users(),
			"clo_storage_s3_user_keys": dataSourceS3Keys(),
			"clo_network_vrouters":     dataSourceVrouters(),
		},
	}
}

func configureProvider(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	bu := data.Get("auth_url").(string)
	at := data.Get("token").(string)
	if len(bu) == 0 {
		return nil, diag.FromErr(errors.New("CLO_API_AUTH_URL parameter should be provided"))
	}
	if len(at) == 0 {
		return nil, diag.FromErr(errors.New("CLO_API_AUTH_TOKEN parameter should be provided"))
	}
	v3cli, e := cloapi.New(at, bu)
	if e != nil {
		return nil, diag.FromErr(e)
	}
	return &providerMeta{v3: v3cli}, nil
}
