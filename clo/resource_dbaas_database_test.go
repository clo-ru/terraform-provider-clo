package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const dbaasDatabaseName = "db_1"

func TestAccCloDbaasDatabase_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Fatal("Error get test client ", err)
	}
	clusterID, err := buildTestDbaasCluster(cli, t)
	if err != nil {
		t.Fatal("Error while create dbaas cluster ", err)
	}

	db := new(cloapi.Database)
	addr := fmt.Sprintf("clo_dbaas_database.%s", dbaasDatabaseName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDbaasDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloDbaasDatabaseConfig(clusterID, "S3cret-pass-1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbaasDatabaseExists(addr, db),
					resource.TestCheckResourceAttr(addr, "name", "appdb"),
					resource.TestCheckResourceAttr(addr, "admin_username", "app_admin"),
					resource.TestCheckResourceAttrSet(addr, "status"),
				),
			},
			{
				// Rotate the admin password via the restore path.
				Config: testAccCloDbaasDatabaseConfig(clusterID, "S3cret-pass-2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbaasDatabaseExists(addr, db),
					resource.TestCheckResourceAttr(addr, "admin_password", "S3cret-pass-2"),
				),
			},
		},
	})
}

func testAccCloDbaasDatabaseConfig(clusterID, password string) string {
	return fmt.Sprintf(`resource "clo_dbaas_database" "%s" {
	cluster_id     = "%s"
	name           = "appdb"
	admin_username = "app_admin"
	admin_password = "%s"
}`, dbaasDatabaseName, clusterID, password)
}

func testAccCheckDbaasDatabaseExists(n string, item *cloapi.Database) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("dbaas database ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		db, e := cli.GetDatabase(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *db
		return nil
	}
}

func testAccCheckDbaasDatabaseDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_dbaas_database" {
			continue
		}
		_, e := cli.GetDatabase(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("dbaas database %s still exists", rs.Primary.ID)
	}
	return nil
}
