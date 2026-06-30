package clo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	creatingVolume  = "CREATING"
	activeVolume    = "AVAILABLE"
	resizingVolume  = "RESIZING"
	attachingVolume = "ATTACHING"
	attachedVolume  = "IN_USE"
	detachingVolume = "DETACHING"
	deletingVolume  = "DELETING"
	deletedVolume   = "DELETED"
)

func resourceVolume() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a new volume in the project",
		ReadContext:   resourceVolumeRead,
		CreateContext: resourceVolumeCreate,
		UpdateContext: resourceVolumeUpdate,
		DeleteContext: resourceVolumeDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("size", func(ctx context.Context, oldValue, newValue, meta interface{}) error {
				if newValue.(int) < oldValue.(int) {
					return fmt.Errorf("size could be increased only")
				}
				return nil
			})),
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the volume should be created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Human-readable name of the new volume",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"size": {
				Description: "Size of the new volume in Gb",
				Type:        schema.TypeInt,
				Required:    true,
				ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
					if sz := i.(int); sz < 10 {
						errs = append(errs, fmt.Errorf("size should be at least 10Gb"))
					}
					return
				},
			},
			"id": {
				Description: "ID of the new volume",
				Type:        schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"created_in": {
				Description: "Timestamp the volume was created",
				Type:        schema.TypeString, Computed: true},
		},
	}
}

func resourceVolumeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	id, err := createVolume(ctx, cli, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := waitVolumeState(ctx, id, cli, []string{creatingVolume}, []string{activeVolume}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)
	return resourceVolumeRead(ctx, d, m)
}

func resourceVolumeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if d.HasChange("size") {
		_, c := d.GetChange("size")
		if err := resizeVolume(ctx, cli, d, c.(int)); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceVolumeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteVolume(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitVolumeDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVolumeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	vol, err := cli.GetVolume(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("id", vol.ID); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", vol.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("created_in", vol.CreatedIn); e != nil {
		return diag.FromErr(e)
	}
	return nil
}

// Waiters
func waitVolumeState(ctx context.Context, id string, cli *cloapi.Client, pending []string, target []string, timeout time.Duration) error {
	return waitForState(ctx, timeout, pending, target, func() (interface{}, string, error) {
		vol, err := cli.GetVolume(ctx, id)
		if err != nil {
			return nil, "", err
		}
		return vol, strings.ToUpper(vol.Status), nil
	})
}

func waitVolumeDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	return waitForState(ctx, timeout, []string{deletingVolume}, []string{deletedVolume}, func() (interface{}, string, error) {
		vol, err := cli.GetVolume(ctx, id)
		if cloapi.IsNotFound(err) {
			return struct{}{}, deletedVolume, nil
		}
		if err != nil {
			return nil, "", err
		}
		return vol, strings.ToUpper(vol.Status), nil
	})
}

// Api actions
func createVolume(ctx context.Context, cli *cloapi.Client, d *schema.ResourceData) (string, error) {
	p := cloapi.VolumeCreateParams{
		ProjectID: d.Get("project_id").(string),
		Size:      d.Get("size").(int),
	}
	if n, ok := d.GetOk("name"); ok {
		p.Name = n.(string)
	} else {
		p.Autorename = true
	}
	return cli.CreateVolume(ctx, p)
}

func resizeVolume(ctx context.Context, cli *cloapi.Client, d *schema.ResourceData, size int) error {
	if err := cli.ExtendVolume(ctx, d.Id(), size); err != nil {
		return err
	}
	return waitVolumeState(ctx, d.Id(), cli, []string{resizingVolume}, []string{activeVolume, attachedVolume}, d.Timeout(schema.TimeoutCreate))
}
