package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const snapshotName = "snapshot_1"

func TestAccCloSnapshot_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Fatal("Error get test client ", err)
	}
	serverID, err := buildTestServer(cli, t)
	if err != nil {
		t.Fatal("Error while create server ", err)
	}

	snap := new(cloapi.Snapshot)
	addr := fmt.Sprintf("clo_compute_snapshot.%s", snapshotName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloSnapshotConfig(serverID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(addr, snap),
					resource.TestCheckResourceAttr(addr, "name", "tf-acc-snapshot"),
					resource.TestCheckResourceAttr(addr, "parent_server", serverID),
					resource.TestCheckResourceAttrSet(addr, "status"),
					resource.TestCheckResourceAttrSet(addr, "deleted_in"),
				),
			},
		},
	})
}

func testAccCloSnapshotConfig(serverID string) string {
	return fmt.Sprintf(`resource "clo_compute_snapshot" "%s" {
	server_id = "%s"
	name      = "tf-acc-snapshot"
}`, snapshotName, serverID)
}

func testAccCheckSnapshotExists(n string, item *cloapi.Snapshot) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("snapshot ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		s, e := cli.GetSnapshot(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *s
		return nil
	}
}

func testAccCheckSnapshotDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_compute_snapshot" {
			continue
		}
		_, e := cli.GetSnapshot(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("snapshot %s still exists", rs.Primary.ID)
	}
	return nil
}
