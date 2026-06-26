package clo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	s3UserActive   = "AVAILABLE"
	s3UserDelete   = "DELETE"
	s3UserCreating = "CREATING"
	s3UserDeleting = "DELETING"
)

func resourceS3User() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a new user of the object storage",
		ReadContext:   resourceS3UserRead,
		CreateContext: resourceS3UserCreate,
		UpdateContext: resourceS3UserUpdate,
		DeleteContext: resourceS3UserDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the user should be created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"canonical_name": {
				Description: "Canonical name of the user. The storage uses this name. Should be unique in scope of the tenant",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"default_bucket": {
				Description: "Should the default bucket be created with the user",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
			"max_buckets": {
				Description: "How many buckets the user could create",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"name": {
				Description: "Human-readable name of the user",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"user_quota_max_size": {
				Description: "Total size of the objects the user can store",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"user_quota_max_objects": {
				Description: "How many objects the user can create",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"bucket_quota_max_size": {
				Description: "A maximum size of a bucket",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"bucket_quota_max_objects": {
				Description: "How many objects can be created within a bucket",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"user_id": {
				Description: "ID of the created user",
				Type:        schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"tenant": {
				Description: "Name of the user's tenant. Name of the user's project uses by default",
				Type:        schema.TypeString, Computed: true},
		},
	}
}

func resourceS3UserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3

	id, err := cli.CreateS3User(ctx, buildS3UserCreateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(id)

	if err := waitS3UserState(ctx, id, cli, []string{s3UserCreating}, []string{s3UserActive}, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	return resourceS3UserRead(ctx, d, m)
}

func buildS3UserCreateParams(d *schema.ResourceData) cloapi.S3UserCreateParams {
	return cloapi.S3UserCreateParams{
		ProjectID:             d.Get("project_id").(string),
		Name:                  optString(d, "name"),
		CanonicalName:         d.Get("canonical_name").(string),
		DefaultBucket:         d.Get("default_bucket").(bool),
		MaxBuckets:            d.Get("max_buckets").(int),
		UserQuotaMaxSize:      d.Get("user_quota_max_size").(int),
		UserQuotaMaxObjects:   d.Get("user_quota_max_objects").(int),
		BucketQuotaMaxSize:    d.Get("bucket_quota_max_size").(int),
		BucketQuotaMaxObjects: d.Get("bucket_quota_max_objects").(int),
	}
}

func resourceS3UserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	uId := d.Id()

	if d.HasChange("name") {
		if err := cli.UpdateS3UserName(ctx, uId, d.Get("name").(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChanges(
		"max_buckets",
		"user_quota_max_size",
		"user_quota_max_objects",
		"bucket_quota_max_size",
		"bucket_quota_max_objects",
	) {
		err := cli.UpdateS3UserQuota(ctx, uId, cloapi.S3UserQuotaParams{
			MaxBuckets:            d.Get("max_buckets").(int),
			UserQuotaMaxSize:      d.Get("user_quota_max_size").(int),
			UserQuotaMaxObjects:   d.Get("user_quota_max_objects").(int),
			BucketQuotaMaxSize:    d.Get("bucket_quota_max_size").(int),
			BucketQuotaMaxObjects: d.Get("bucket_quota_max_objects").(int),
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return resourceS3UserRead(ctx, d, m)
}

func resourceS3UserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	user, err := cli.GetS3User(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if e := d.Set("user_id", user.ID); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", user.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("tenant", user.Tenant); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("max_buckets", user.MaxBuckets); e != nil {
		return diag.FromErr(e)
	}
	return dispatchQuotaInfo(d, user.Quotas)
}

func dispatchQuotaInfo(d *schema.ResourceData, q []cloapi.S3Quota) diag.Diagnostics {
	for _, qi := range q {
		switch qi.Type {
		case "user":
			if e := d.Set("user_quota_max_size", qi.MaxSize); e != nil {
				return diag.FromErr(e)
			}
			if e := d.Set("user_quota_max_objects", qi.MaxObjects); e != nil {
				return diag.FromErr(e)
			}
		case "bucket":
			if e := d.Set("bucket_quota_max_size", qi.MaxSize); e != nil {
				return diag.FromErr(e)
			}
			if e := d.Set("bucket_quota_max_objects", qi.MaxObjects); e != nil {
				return diag.FromErr(e)
			}
		default:
			return diag.FromErr(fmt.Errorf("wrong the quota_info type, %s", qi.Type))
		}
	}
	return nil
}

func resourceS3UserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*providerMeta).v3
	if err := cli.DeleteS3User(ctx, d.Id()); err != nil {
		return diag.FromErr(err)
	}
	if err := waitS3UserDeleted(ctx, d.Id(), cli, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// waiters

func waitS3UserState(ctx context.Context, id string, cli *cloapi.Client, pending []string, target []string, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			user, err := cli.GetS3User(ctx, id)
			if err != nil {
				return nil, "", err
			}
			return user, user.Status, nil
		},
		Pending:    pending,
		Target:     target,
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, stateConf.Timeout, func() *resource.RetryError {
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}

func waitS3UserDeleted(ctx context.Context, id string, cli *cloapi.Client, timeout time.Duration) error {
	stateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			user, err := cli.GetS3User(ctx, id)
			if cloapi.IsNotFound(err) {
				return struct{}{}, s3UserDelete, nil
			}
			if err != nil {
				return nil, "", err
			}
			return user, user.Status, nil
		},
		Pending:    []string{s3UserDeleting},
		Target:     []string{s3UserDelete},
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 30 * time.Second,
	}
	return resource.RetryContext(ctx, stateConf.Timeout, func() *resource.RetryError {
		if _, err := stateConf.WaitForStateContext(ctx); err != nil {
			log.Printf("[DEBUG] Retrying after error: %s", err)
			return &resource.RetryError{Err: err}
		}
		return nil
	})
}
