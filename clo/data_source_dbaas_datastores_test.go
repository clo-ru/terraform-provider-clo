package clo

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCloDbaasDatastoresDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloDbaasDatastoresDataSource(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.clo_dbaas_datastores.all", "result.#"),
				),
			},
		},
	})
}

func testAccCloDbaasDatastoresDataSource() string {
	return fmt.Sprintf(`data "clo_dbaas_datastores" "all"{
			project_id = "%s"
	}`, projectID)
}
