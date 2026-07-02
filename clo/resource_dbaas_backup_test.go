package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const dbaasBackupName = "backup_1"

func TestAccCloDbaasBackup_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Fatal("Error get test client ", err)
	}
	clusterID, err := buildTestDbaasCluster(cli, t)
	if err != nil {
		t.Fatal("Error while create dbaas cluster ", err)
	}

	bk := new(cloapi.Backup)
	addr := fmt.Sprintf("clo_dbaas_backup.%s", dbaasBackupName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDbaasBackupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloDbaasBackupConfig(clusterID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbaasBackupExists(addr, bk),
					resource.TestCheckResourceAttr(addr, "name", "tf-acc-backup"),
					resource.TestCheckResourceAttr(addr, "type", "FULL"),
					resource.TestCheckResourceAttrSet(addr, "status"),
				),
			},
		},
	})
}

func testAccCloDbaasBackupConfig(clusterID string) string {
	return fmt.Sprintf(`resource "clo_dbaas_backup" "%s" {
	cluster_id   = "%s"
	name         = "tf-acc-backup"
	force_delete = true
}`, dbaasBackupName, clusterID)
}

func testAccCheckDbaasBackupExists(n string, item *cloapi.Backup) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("dbaas backup ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		b, e := cli.GetBackup(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *b
		return nil
	}
}

func testAccCheckDbaasBackupDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_dbaas_backup" {
			continue
		}
		_, e := cli.GetBackup(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("dbaas backup %s still exists", rs.Primary.ID)
	}
	return nil
}
