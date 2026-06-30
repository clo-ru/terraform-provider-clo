package clo

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// waitForState polls refresh until the resource reaches one of the target
// states. It centralizes the StateChangeConf timing shared by every resource
// waiter.
func waitForState(ctx context.Context, timeout time.Duration, pending, target []string, refresh resource.StateRefreshFunc) error {
	stateConf := resource.StateChangeConf{
		Refresh:    refresh,
		Pending:    pending,
		Target:     target,
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	if _, err := stateConf.WaitForStateContext(ctx); err != nil {
		log.Printf("[DEBUG] wait for state failed: %s", err)
		return err
	}
	return nil
}