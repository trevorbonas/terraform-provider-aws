// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_addon", name="Add-On")
// @Tags(identifierAttribute="arn")
func resourceAddon() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAddonCreate,
		ReadWithoutTimeout:   resourceAddonRead,
		UpdateWithoutTimeout: resourceAddonUpdate,
		DeleteWithoutTimeout: resourceAddonDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(40 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"addon_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					// Regular expression taken from: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
					validation.StringMatch(regexache.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9]\d*|\d*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`), "must follow semantic version format"),
				),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"configuration_values": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pod_identity_association": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"service_account": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"preserve": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"resolve_conflicts_on_create": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice(enum.Slice(
					types.ResolveConflictsNone,
					types.ResolveConflictsOverwrite,
				), false),
			},
			"resolve_conflicts_on_update": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.ResolveConflicts](),
			},
			"service_account_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAddonCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get(names.AttrClusterName).(string)
	id := AddonCreateResourceID(clusterName, addonName)
	input := &eks.CreateAddonInput{
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		ClusterName:        aws.String(clusterName),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("addon_version"); ok {
		input.AddonVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("configuration_values"); ok {
		input.ConfigurationValues = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pod_identity_association"); ok && v.(*schema.Set).Len() > 0 {
		input.PodIdentityAssociations = expandAddonPodIdentityAssociations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("resolve_conflicts_on_create"); ok {
		input.ResolveConflicts = types.ResolveConflicts(v.(string))
	}

	if v, ok := d.GetOk("service_account_role_arn"); ok {
		input.ServiceAccountRoleArn = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateAddon(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "CREATE_FAILED") {
				return true, err
			}

			if errs.IsAErrorMessageContains[*types.InvalidParameterException](err, "does not exist") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EKS Add-On (%s): %s", id, err)
	}

	d.SetId(id)

	if _, err := waitAddonCreated(ctx, conn, clusterName, addonName, d.Timeout(schema.TimeoutCreate)); err != nil {
		// Creating addon w/o setting resolve_conflicts_on_create to "OVERWRITE"
		// might result in a failed creation, if unmanaged version of addon is already deployed
		// and there are configuration conflicts:
		// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
		//
		// Addon resource is tainted after failed creation, thus will be deleted and created again.
		// Re-creating like this will resolve the error, but it will also purge any
		// configurations that were applied by the user (that were conflicting). This might we an unwanted
		// side effect and should be left for the user to decide how to handle it.
		diags = sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) create: %s", d.Id(), err)
		return sdkdiag.AppendWarningf(diags, "Running terraform apply again will remove the kubernetes add-on and attempt to create it again effectively purging previous add-on configuration")
	}

	return append(diags, resourceAddonRead(ctx, d, meta)...)
}

func resourceAddonRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, addonName, err := AddonParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	addon, err := findAddonByTwoPartKey(ctx, conn, clusterName, addonName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Add-On (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Add-On (%s): %s", d.Id(), err)
	}

	d.Set("addon_name", addon.AddonName)
	d.Set("addon_version", addon.AddonVersion)
	d.Set(names.AttrARN, addon.AddonArn)
	d.Set(names.AttrClusterName, addon.ClusterName)
	d.Set("configuration_values", addon.ConfigurationValues)
	d.Set(names.AttrCreatedAt, aws.ToTime(addon.CreatedAt).Format(time.RFC3339))
	d.Set("modified_at", aws.ToTime(addon.ModifiedAt).Format(time.RFC3339))
	flatPIAs, err := flattenAddonPodIdentityAssociations(ctx, addon.PodIdentityAssociations, clusterName, meta)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "flattening pod_identity_association: %s", err)
	}
	if err := d.Set("pod_identity_association", flatPIAs); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting pod_identity_association: %s", err)
	}
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)

	setTagsOut(ctx, addon.Tags)

	return diags
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, addonName, err := AddonParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChanges("addon_version", "service_account_role_arn", "configuration_values", "pod_identity_association") {
		input := &eks.UpdateAddonInput{
			AddonName:          aws.String(addonName),
			ClientRequestToken: aws.String(sdkid.UniqueId()),
			ClusterName:        aws.String(clusterName),
		}

		if d.HasChange("addon_version") {
			input.AddonVersion = aws.String(d.Get("addon_version").(string))
		}

		if d.HasChange("configuration_values") {
			input.ConfigurationValues = aws.String(d.Get("configuration_values").(string))
		}

		if d.HasChange("pod_identity_association") {
			if v, ok := d.GetOk("pod_identity_association"); ok && v.(*schema.Set).Len() > 0 {
				input.PodIdentityAssociations = expandAddonPodIdentityAssociations(v.(*schema.Set).List())
			} else {
				input.PodIdentityAssociations = []types.AddonPodIdentityAssociations{}
			}
		}

		if v, ok := d.GetOk("resolve_conflicts_on_update"); ok {
			input.ResolveConflicts = types.ResolveConflicts(v.(string))
		}

		// If service account role ARN is already provided, use it. Otherwise, the add-on uses
		// permissions assigned to the node IAM role.
		if d.HasChange("service_account_role_arn") || d.Get("service_account_role_arn").(string) != "" {
			input.ServiceAccountRoleArn = aws.String(d.Get("service_account_role_arn").(string))
		}

		output, err := conn.UpdateAddon(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EKS Add-On (%s): %s", d.Id(), err)
		}

		updateID := aws.ToString(output.Update.Id)
		if _, err := waitAddonUpdateSuccessful(ctx, conn, clusterName, addonName, updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			if input.ResolveConflicts != types.ResolveConflictsOverwrite {
				// Changing addon version w/o setting resolve_conflicts_on_update to "OVERWRITE"
				// might result in a failed update if there are conflicts:
				// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
				return sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) update (%s): %s. Consider setting resolve_conflicts_on_update to %q", d.Id(), updateID, err, types.ResolveConflictsOverwrite)
			}

			return sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) update (%s): %s", d.Id(), updateID, err)
		}
	}

	return append(diags, resourceAddonRead(ctx, d, meta)...)
}

func resourceAddonDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, addonName, err := AddonParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &eks.DeleteAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	if v, ok := d.GetOk("preserve"); ok {
		input.Preserve = v.(bool)
	}

	log.Printf("[DEBUG] Deleting EKS Add-On: %s", d.Id())
	_, err = conn.DeleteAddon(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EKS Add-On (%s): %s", d.Id(), err)
	}

	if _, err := waitAddonDeleted(ctx, conn, clusterName, addonName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EKS Add-On (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandAddonPodIdentityAssociations(tfList []any) []types.AddonPodIdentityAssociations {
	if len(tfList) == 0 {
		return nil
	}

	var addonPodIdentityAssociations []types.AddonPodIdentityAssociations
	for _, raw := range tfList {
		tfMap, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		pia := types.AddonPodIdentityAssociations{}
		if roleArn, ok := tfMap[names.AttrRoleARN].(string); ok {
			pia.RoleArn = aws.String(roleArn)
		}
		if service_account, ok := tfMap["service_account"].(string); ok {
			pia.ServiceAccount = aws.String(service_account)
		}

		addonPodIdentityAssociations = append(addonPodIdentityAssociations, pia)
	}

	return addonPodIdentityAssociations
}

func flattenAddonPodIdentityAssociations(ctx context.Context, associations []string, clusterName string, meta any) ([]any, error) {
	if len(associations) == 0 {
		return nil, nil
	}

	conn := meta.(*conns.AWSClient).EKSClient(ctx)
	var results []any

	for _, associationArn := range associations {
		// CreateAddon returns the associationARN. The associationId is extracted from the end of the ARN,
		// which is used in the DescribePodIdentityAssociation call to get the RoleARN and ServiceAccount
		//
		// Ex. "arn:aws:eks:<region>:<account-id>:podidentityassociation/<cluster-name>/a-1v95i5dqqiylbo3ud"
		parts := strings.Split(associationArn, "/")
		if len(parts) != 3 {
			return nil, fmt.Errorf(`unable to extract association ID from ARN "%s"`, associationArn)
		}

		associationId := parts[2]
		pia, err := findPodIdentityAssociationByTwoPartKey(ctx, conn, associationId, clusterName)
		if err != nil {
			return nil, err
		}

		tfMap := map[string]any{
			names.AttrRoleARN: pia.RoleArn,
			"service_account": pia.ServiceAccount,
		}

		results = append(results, tfMap)
	}

	return results, nil
}

func findAddonByTwoPartKey(ctx context.Context, conn *eks.Client, clusterName, addonName string) (*types.Addon, error) {
	input := &eks.DescribeAddonInput{
		AddonName:   aws.String(addonName),
		ClusterName: aws.String(clusterName),
	}

	output, err := conn.DescribeAddon(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Addon == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Addon, nil
}

func findAddonUpdateByThreePartKey(ctx context.Context, conn *eks.Client, clusterName, addonName, id string) (*types.Update, error) {
	input := &eks.DescribeUpdateInput{
		AddonName: aws.String(addonName),
		Name:      aws.String(clusterName),
		UpdateId:  aws.String(id),
	}

	output, err := conn.DescribeUpdate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Update == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Update, nil
}

func statusAddon(ctx context.Context, conn *eks.Client, clusterName, addonName string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findAddonByTwoPartKey(ctx, conn, clusterName, addonName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusAddonUpdate(ctx context.Context, conn *eks.Client, clusterName, addonName, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findAddonUpdateByThreePartKey(ctx, conn, clusterName, addonName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitAddonCreated(ctx context.Context, conn *eks.Client, clusterName, addonName string, timeout time.Duration) (*types.Addon, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(types.AddonStatusCreating, types.AddonStatusDegraded),
		Target:  enum.Slice(types.AddonStatusActive),
		Refresh: statusAddon(ctx, conn, clusterName, addonName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Addon); ok {
		if status, health := output.Status, output.Health; status == types.AddonStatusCreateFailed && health != nil {
			tfresource.SetLastError(err, addonIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitAddonDeleted(ctx context.Context, conn *eks.Client, clusterName, addonName string, timeout time.Duration) (*types.Addon, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AddonStatusActive, types.AddonStatusDeleting),
		Target:  []string{},
		Refresh: statusAddon(ctx, conn, clusterName, addonName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Addon); ok {
		if status, health := output.Status, output.Health; status == types.AddonStatusDeleteFailed && health != nil {
			tfresource.SetLastError(err, addonIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitAddonUpdateSuccessful(ctx context.Context, conn *eks.Client, clusterName, addonName, id string, timeout time.Duration) (*types.Update, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(types.UpdateStatusInProgress),
		Target:  enum.Slice(types.UpdateStatusSuccessful),
		Refresh: statusAddonUpdate(ctx, conn, clusterName, addonName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Update); ok {
		if status := output.Status; status == types.UpdateStatusCancelled || status == types.UpdateStatusFailed {
			tfresource.SetLastError(err, errorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func addonIssueError(apiObject types.AddonIssue) error {
	return fmt.Errorf("%s: %s", apiObject.Code, aws.ToString(apiObject.Message))
}

func addonIssuesError(apiObjects []types.AddonIssue) error {
	var errs []error

	for _, apiObject := range apiObjects {
		err := addonIssueError(apiObject)

		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", strings.Join(apiObject.ResourceIds, ", "), err))
		}
	}

	return errors.Join(errs...)
}

func errorDetailError(apiObject types.ErrorDetail) error {
	return fmt.Errorf("%s: %s", apiObject.ErrorCode, aws.ToString(apiObject.ErrorMessage))
}

func errorDetailsError(apiObjects []types.ErrorDetail) error {
	var errs []error

	for _, apiObject := range apiObjects {
		err := errorDetailError(apiObject)

		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", strings.Join(apiObject.ResourceIds, ", "), err))
		}
	}

	return errors.Join(errs...)
}
