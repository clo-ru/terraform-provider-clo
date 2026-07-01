package clo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRecipe() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches a recipe by name. Useful to resolve a recipe_id for the instance resource.",
		ReadContext: dataSourceRecipeRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns recipes",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The name of the desired recipe",
				Type:        schema.TypeString,
				Required:    true,
			},
			"recipe_id": {
				Description: "ID of the requested recipe",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"min_disk": {
				Description: "Minimum disk size (Gb) required by the recipe",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"min_ram": {
				Description: "Minimum RAM (Gb) required by the recipe",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"min_vcpus": {
				Description: "Minimum number of vCPUs required by the recipe",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"suitable_images": {
				Description: "IDs of images the recipe can be applied to",
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceRecipeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	recipes, err := cli.ListRecipes(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	name := d.Get("name").(string)
	for _, r := range recipes {
		if r.Name != name {
			continue
		}
		images := make([]interface{}, 0, len(r.SuitableImages))
		for _, im := range r.SuitableImages {
			images = append(images, im)
		}
		fields := map[string]interface{}{
			"recipe_id":       r.ID,
			"min_disk":        r.MinDisk,
			"min_ram":         r.MinRam,
			"min_vcpus":       r.MinVcpus,
			"suitable_images": images,
		}
		for k, v := range fields {
			if e := d.Set(k, v); e != nil {
				return diag.FromErr(e)
			}
		}
		d.SetId(r.ID)
		break
	}
	return nil
}
