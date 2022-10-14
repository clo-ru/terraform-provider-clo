package clo

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

const (
	dsS3UserName = "s3user"
)

func TestAccCloS3UserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloS3UserDataSourceBasic(),
			},
			{
				Config: testAccCloS3UserDataSourceSource(),
				Check: resource.ComposeTestCheckFunc(
					testAccCloS3UserDataSourceID("data.clo_storage_s3_user.source"),
				),
			},
			{
				Config: testAccCloS3UserKeysDataSourceBasic(),
				Check:  testAccCloS3UserKeysDataSourceAccess("clo_storage_s3_user_keys.source_key"),
			},
		},
	})
}

func testAccCloS3UserDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find s3User data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("s3User data source ID not set")
		}

		return nil
	}
}

func testAccCloS3UserDataSourceBasic() string {
	return fmt.Sprintf(`resource "clo_storage_s3_user" "%s"{
 			max_buckets = 2
 			project_id = "%s"
 			canonical_name = "%s"
		 	user_quota_max_size=30
	}`, dsS3UserName, projectID, dsS3UserName)
}

func testAccCloS3UserKeysDataSourceAccess(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find s3UserKey data source: %s", n)
		}
		return nil
	}
}

func testAccCloS3UserKeysDataSourceBasic() string {
	return fmt.Sprintf(`
	%s

	resource "clo_storage_s3_user_keys" "source_key"{
			user_id = "${clo_storage_s3_user.%s.id}"
	}`, testAccCloS3UserDataSourceBasic(), dsS3UserName)
}

func testAccCloS3UserDataSourceSource() string {
	return fmt.Sprintf(`
		%s
		data "clo_storage_s3_user" "source" {
			user_id = "${clo_storage_s3_user.%s.id}"
		}`, testAccCloS3UserDataSourceBasic(), dsS3UserName,
	)
}
