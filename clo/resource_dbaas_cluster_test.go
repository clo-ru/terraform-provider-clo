package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const dbaasClusterName = "cluster_1"

func TestAccCloDbaasCluster_basic(t *testing.T) {
	cl := new(cloapi.Cluster)
	addr := fmt.Sprintf("clo_dbaas_cluster.%s", dbaasClusterName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckDbaasClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloDbaasClusterConfig(dbaasClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbaasClusterExists(addr, cl),
					resource.TestCheckResourceAttr(addr, "enabled", "true"),
					resource.TestCheckResourceAttr(addr, "storage_size", "10"),
					resource.TestCheckResourceAttrSet(addr, "status"),
					resource.TestCheckResourceAttrSet(addr, "datastore_id"),
				),
			},
			{
				Config: testAccCloDbaasClusterConfig("cluster_renamed"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDbaasClusterExists(addr, cl),
					resource.TestCheckResourceAttr(addr, "name", "cluster_renamed"),
				),
			},
		},
	})
}

func testAccCloDbaasClusterConfig(name string) string {
	return fmt.Sprintf(`
data "clo_dbaas_datastores" "all" {
	project_id = "%s"
}

resource "clo_dbaas_cluster" "%s" {
	project_id   = "%s"
	name         = "%s"
	datastore_id = data.clo_dbaas_datastores.all.result[0].id
	storage_size = 10
	flavor {
		vcpus = 1
		ram   = 2
	}
}`, projectID, dbaasClusterName, projectID, name)
}

func testAccCheckDbaasClusterExists(n string, item *cloapi.Cluster) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("dbaas cluster ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		cl, e := cli.GetCluster(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *cl
		return nil
	}
}

func testAccCheckDbaasClusterDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_dbaas_cluster" {
			continue
		}
		_, e := cli.GetCluster(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("dbaas cluster %s still exists", rs.Primary.ID)
	}
	return nil
}
