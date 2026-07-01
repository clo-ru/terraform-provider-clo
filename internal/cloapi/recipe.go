package cloapi

import (
	"context"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Recipe is the provider-facing view of a project recipe. A recipe bundles the
// minimum flavor (disk/ram/vcpus) with the images and licenses it can be applied
// to. The instance resource accepts a recipe_id, so this lets users resolve one
// by name instead of hard-coding an ID.
type Recipe struct {
	ID             string
	Name           string
	MinDisk        int
	MinRam         int
	MinVcpus       int
	SuitableImages []string
	Licenses       []RecipeLicense
}

// RecipeLicense is a license offered by a recipe.
type RecipeLicense struct {
	Addon    string
	Name     string
	Required bool
}

func recipeFromSchema(r *gen.SchemasResponseV2RecipeRecipeSchema) Recipe {
	rec := Recipe{
		ID:             r.Id,
		Name:           r.Name,
		MinDisk:        r.MinDisk,
		MinRam:         r.MinRam,
		MinVcpus:       r.MinVcpus,
		SuitableImages: append([]string(nil), r.SuitableImages...),
	}
	if r.AvailableLicenses != nil {
		for _, l := range *r.AvailableLicenses {
			rl := RecipeLicense{Addon: l.Addon, Required: l.Required}
			if l.Name != nil {
				rl.Name = *l.Name
			}
			rec.Licenses = append(rec.Licenses, rl)
		}
	}
	return rec
}

// ListRecipes returns the project's recipes (single page, matching the other list adapters).
func (c *Client) ListRecipes(ctx context.Context, projectID string) ([]Recipe, error) {
	resp, err := c.gen.ProjectRecipesWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Recipe, 0, len(items))
	for i := range items {
		out = append(out, recipeFromSchema(&items[i]))
	}
	return out, nil
}
