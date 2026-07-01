package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	loadBalancerName     = "lb_1"
	loadBalancerRuleName = "rule_1"
)

func TestAccCloLoadBalancer_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Fatal("Error get test client ", err)
	}
	addrID, err := buildTestAddress(cli, t)
	if err != nil {
		t.Fatal("Error while create address ", err)
	}

	lb := new(cloapi.LoadBalancer)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloLoadBalancerBasic(addrID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerExists(fmt.Sprintf("clo_network_loadbalancer.%s", loadBalancerName), lb),
					resource.TestCheckResourceAttr(fmt.Sprintf("clo_network_loadbalancer.%s", loadBalancerName), "enabled", "true"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("clo_network_loadbalancer.%s", loadBalancerName), "status"),
				),
			},
		},
	})
}

func TestAccCloLoadBalancerRule_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Fatal("Error get test client ", err)
	}
	// The LB needs an external address; a rule targets a backend server's
	// internal (FIXED) address, so build both fixtures.
	lbAddrID, err := buildTestAddress(cli, t)
	if err != nil {
		t.Fatal("Error while create address ", err)
	}
	serverID, err := buildTestServer(cli, t)
	if err != nil {
		t.Fatal("Error while create server ", err)
	}
	srv, err := cli.GetServer(context.Background(), serverID)
	if err != nil {
		t.Fatal("Error get server ", err)
	}
	if len(srv.Addresses) == 0 {
		t.Fatal("server has no address to target with a rule")
	}
	backendAddrID := srv.Addresses[0]

	rule := new(cloapi.Rule)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckLoadBalancerRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloLoadBalancerRuleBasic(lbAddrID, backendAddrID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLoadBalancerRuleExists(fmt.Sprintf("clo_network_loadbalancer_rule.%s", loadBalancerRuleName), rule),
					resource.TestCheckResourceAttr(fmt.Sprintf("clo_network_loadbalancer_rule.%s", loadBalancerRuleName), "external_protocol_port", "80"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("clo_network_loadbalancer_rule.%s", loadBalancerRuleName), "status"),
				),
			},
		},
	})
}

func testAccCloLoadBalancerBasic(addrID string) string {
	return fmt.Sprintf(`resource "clo_network_loadbalancer" "%s" {
	project_id = "%s"
	name       = "%s"
	algorithm  = "ROUND_ROBIN"

	address {
		id = "%s"
	}

	healthmonitor {
		type        = "TCP"
		delay       = 80
		timeout     = 15
		max_retries = 3
	}
}`, loadBalancerName, projectID, loadBalancerName, addrID)
}

func testAccCloLoadBalancerRuleBasic(lbAddrID, backendAddrID string) string {
	return fmt.Sprintf(`resource "clo_network_loadbalancer" "%s" {
	project_id = "%s"
	name       = "%s"

	address {
		id = "%s"
	}

	healthmonitor {
		type        = "TCP"
		delay       = 80
		timeout     = 15
		max_retries = 3
	}
}

resource "clo_network_loadbalancer_rule" "%s" {
	loadbalancer_id        = clo_network_loadbalancer.%s.id
	address_id             = "%s"
	external_protocol_port = 80
	internal_protocol_port = 8080
}`, loadBalancerName, projectID, loadBalancerName, lbAddrID, loadBalancerRuleName, loadBalancerName, backendAddrID)
}

func testAccCheckLoadBalancerExists(n string, item *cloapi.LoadBalancer) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("loadbalancer ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		lb, e := cli.GetLoadBalancer(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *lb
		return nil
	}
}

func testAccCheckLoadBalancerDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_network_loadbalancer" {
			continue
		}
		_, e := cli.GetLoadBalancer(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("loadbalancer %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckLoadBalancerRuleExists(n string, item *cloapi.Rule) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("rule ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		r, e := cli.GetRule(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *r
		return nil
	}
}

func testAccCheckLoadBalancerRuleDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_network_loadbalancer_rule" {
			continue
		}
		_, e := cli.GetRule(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("rule %s still exists", rs.Primary.ID)
	}
	return nil
}
