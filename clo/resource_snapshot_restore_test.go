package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const snapshotRestoreName = "restore_1"

func TestAccCloSnapshotRestore_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Fatal("Error get test client ", err)
	}
	snapshotID, err := buildTestSnapshot(cli, t)
	if err != nil {
		t.Fatal("Error while create snapshot ", err)
	}

	srv := new(cloapi.Server)
	addr := fmt.Sprintf("clo_compute_snapshot_restore.%s", snapshotRestoreName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSnapshotRestoreDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloSnapshotRestoreConfig(snapshotID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotRestoreExists(addr, srv),
					resource.TestCheckResourceAttr(addr, "name", "tf-acc-restored"),
					resource.TestCheckResourceAttrSet(addr, "status"),
				),
			},
		},
	})
}

func testAccCloSnapshotRestoreConfig(snapshotID string) string {
	return fmt.Sprintf(`resource "clo_compute_snapshot_restore" "%s" {
	snapshot_id = "%s"
	name        = "tf-acc-restored"
}`, snapshotRestoreName, snapshotID)
}

func testAccCheckSnapshotRestoreExists(n string, item *cloapi.Server) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("restored server ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		srv, e := cli.GetServer(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *srv
		return nil
	}
}

func testAccCheckSnapshotRestoreDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_compute_snapshot_restore" {
			continue
		}
		_, e := cli.GetServer(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("restored server %s still exists", rs.Primary.ID)
	}
	return nil
}
