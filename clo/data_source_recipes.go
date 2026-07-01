package clo

import (
	"context"
	"strconv"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRecipes() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the list of recipes available in the project",
		ReadContext: dataSourceRecipesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project that owns recipes",
				Type:        schema.TypeString,
				Required:    true,
			},
			"result": {
				Description: "The object that holds the results",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: recipeElemSchema(),
				},
			},
		},
	}
}

// recipeElemSchema is the per-recipe attribute set shared by the list result and
// the single-recipe data source.
func recipeElemSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Description: "ID of the recipe",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Name of the recipe",
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
		"available_licenses": {
			Description: "Licenses offered by the recipe",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"addon":    {Type: schema.TypeString, Computed: true},
					"name":     {Type: schema.TypeString, Computed: true},
					"required": {Type: schema.TypeBool, Computed: true},
				},
			},
		},
	}
}

func flattenRecipe(r cloapi.Recipe) map[string]interface{} {
	licenses := make([]interface{}, 0, len(r.Licenses))
	for _, l := range r.Licenses {
		licenses = append(licenses, map[string]interface{}{
			"addon":    l.Addon,
			"name":     l.Name,
			"required": l.Required,
		})
	}
	images := make([]interface{}, 0, len(r.SuitableImages))
	for _, im := range r.SuitableImages {
		images = append(images, im)
	}
	return map[string]interface{}{
		"id":                 r.ID,
		"name":               r.Name,
		"min_disk":           r.MinDisk,
		"min_ram":            r.MinRam,
		"min_vcpus":          r.MinVcpus,
		"suitable_images":    images,
		"available_licenses": licenses,
	}
}

func dataSourceRecipesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	recipes, err := cli.ListRecipes(ctx, d.Get("project_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	res := make([]interface{}, 0, len(recipes))
	for _, r := range recipes {
		res = append(res, flattenRecipe(r))
	}
	if e := d.Set("result", res); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return nil
}
