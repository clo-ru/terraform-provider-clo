package clo

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	dsServerName = "serv"
)

func TestAccCloInstanceDataSource(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Error("Error get test client ", err)
	}
	imageID := getTestImageID(t, cli)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloInstanceDataSourceBasic(imageID),
			},
			{
				Config: testAccCloInstanceDataSourceSource(imageID),
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

func testAccCloInstanceDataSourceBasic(imageID string) string {
	return fmt.Sprintf(`resource "clo_compute_instance" "%s" {
  				project_id = "%s"
  				name = "%s"
  				image_id = "%s"
  				flavor_ram = 4
  				flavor_vcpus = 2
  				block_device{
   					size = 10
   					bootable=true
   					storage_type = "volume"
  				}
  				addresses{
   					version = 4
   					external=true
   					ddos_protection=false
  				}
	}`, dsServerName, projectID, dsServerName, imageID)
}

func testAccCloInstanceDataSourceSource(imageID string) string {
	return fmt.Sprintf(`
		%s

		data "clo_compute_instance" "source_1" {
			id = "${clo_compute_instance.%s.id}"
		}`, testAccCloInstanceDataSourceBasic(imageID), dsServerName,
	)
}
