package clo

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

const (
	dsVolumeName = "volume_1"
)

func TestAccCloVolumeDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloVolumeDataSourceBasic(),
			},
			{
				Config: testAccCloVolumeDataSourceSource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCloVolumeDataSourceID(fmt.Sprintf("clo_disks_volume.%s", dsVolumeName)),
				),
			},
		},
	})
}

func testAccCloVolumeDataSourceBasic() string {
	return fmt.Sprintf(`resource "clo_disks_volume" "%s"{
			project_id = "%s"
			name = "%s"
			size = 30
	}`, dsVolumeName, projectID, dsVolumeName)
}

func testAccCloVolumeDataSourceSource() string {
	return fmt.Sprintf(`
	%s

	data "clo_disks_volume" "%s_1"{
		volume_id = "${clo_disks_volume.%s.id}"
	}`, testAccCloVolumeDataSourceBasic(), dsVolumeName, dsVolumeName)
}

func testAccCloVolumeDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find volume data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("volume data source ID not set")
		}

		return nil
	}
}
