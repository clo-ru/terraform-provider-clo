package clo

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDbaasClusterConfig() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches the tunable configuration of a dbaas cluster: the live (current), datastore/flavor default, and last known-stable parameter sets. All values are read-only and rendered as strings.",
		ReadContext: dataSourceDbaasClusterConfigRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Description: "ID of the dbaas cluster whose configuration is read",
				Type:        schema.TypeString,
				Required:    true,
			},
			"current": {
				Description: "The cluster's live configuration (parameter name -> value)",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"default": {
				Description: "The default configuration for the cluster's datastore and flavor",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"last_stable": {
				Description: "The last known-stable configuration",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDbaasClusterConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id := d.Get("cluster_id").(string)
	cfg, err := cli.GetClusterConfig(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("current", stringifyConfigMap(cfg.Current)); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("default", stringifyConfigMap(cfg.Default)); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("last_stable", stringifyConfigMap(cfg.LastStable)); e != nil {
		return diag.FromErr(e)
	}
	d.SetId(id)
	return nil
}

// stringifyConfigMap renders a heterogeneous config map to string values for a
// TypeMap attribute.
func stringifyConfigMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = stringifyConfigValue(v)
	}
	return out
}

// stringifyConfigValue renders a single config value as text. Scalars keep
// their natural form (integers without a decimal point); composite values
// (lists/objects, e.g. multi-value enums) are JSON-encoded.
func stringifyConfigValue(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		// JSON numbers decode as float64; render whole numbers without a decimal point.
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
