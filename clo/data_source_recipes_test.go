package clo

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloRecipesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloRecipesDataSource(),
				Check: resource.ComposeTestCheckFunc(
					// The project always exposes at least the built-in recipes.
					resource.TestCheckResourceAttrSet("data.clo_project_recipes.all", "result.#"),
				),
			},
		},
	})
}

func testAccCloRecipesDataSource() string {
	return fmt.Sprintf(`data "clo_project_recipes" "all"{
			project_id = "%s"
	}`, projectID)
}
