package clo

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

const (
	dsPtr    = "myPtr.com"
	dsIpName = "serv"
)

func TestAccCloIpDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloIpDataSourceBasic(),
			},
			{
				Config: testAccCloIpDataSourceSource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCloIpDataSourceID("data.clo_network_ip.source_1"),
				),
			},
		},
	})
}

func testAccCloIpDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find ip data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ip data source ID not set")
		}

		return nil
	}
}

func testAccCloIpDataSourceBasic() string {
	return fmt.Sprintf(`resource "clo_network_ip" "%s" {
  				project_id = "%s" 
  				ptr = "%s"
	}`, dsIpName, projectID, dsPtr)
}

func testAccCloIpDataSourceSource() string {
	return fmt.Sprintf(`
		%s
		
		data "clo_network_ip" "source_1" {
			address_id = "${clo_network_ip.%s.id}"
		}`, testAccCloIpDataSourceBasic(), dsIpName,
	)
}
