package clo

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

const (
	dsServerName = "serv"
	dsImageID    = "2dc56270-c4b6-4d2c-b238-8fa58f35634d"
)

func TestAccCloInstanceDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloInstanceDataSourceBasic(),
			},
			{
				Config: testAccCloInstanceDataSourceSource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCloInstanceDataSourceID("data.clo_compute_instance.source_1"),
				),
			},
		},
	})
}

func testAccCloInstanceDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find compute instance data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("compute instance data source ID not set")
		}

		return nil
	}
}

func testAccCloInstanceDataSourceBasic() string {
	return fmt.Sprintf(`resource "clo_compute_instance" "%s" {
  				project_id = "%s" 
  				name = "%s"
  				image_id = "%s"
  				flavor_ram = 4
  				flavor_vcpus = 2
  				block_device{
   					size = 40
   					bootable=true
   					storage_type = "volume"
  				}
  				addresses{
   					version = 4
   					external=true
   					ddos_protection=false
  				}
	}`, dsServerName, projectID, dsServerName, dsImageID)
}

func testAccCloInstanceDataSourceSource() string {
	return fmt.Sprintf(`
		%s
		
		data "clo_compute_instance" "source_1" {
			id = "${clo_compute_instance.%s.id}"
		}`, testAccCloInstanceDataSourceBasic(), dsServerName,
	)
}
