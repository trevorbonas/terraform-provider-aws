// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devopsguru"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devopsguru/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_devopsguru_notification_channel", name="Notification Channel")
func newNotificationChannelResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &notificationChannelResource{}, nil
}

const (
	ResNameNotificationChannel = "Notification Channel"
)

type notificationChannelResource struct {
	framework.ResourceWithModel[notificationChannelResourceModel]
	framework.WithImportByID
}

func (r *notificationChannelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"filters": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[filtersData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"message_types": schema.SetAttribute{
							Optional:    true,
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.NotificationMessageType](),
								),
							},
						},
						"severities": schema.SetAttribute{
							Optional:    true,
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Validators: []validator.Set{
								setvalidator.ValueStringsAre(
									enum.FrameworkValidate[awstypes.InsightSeverity](),
								),
							},
						},
					},
				},
			},
			"sns": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[snsData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrTopicARN: schema.StringAttribute{
							Required:   true,
							CustomType: fwtypes.ARNType,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *notificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var plan notificationChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := &awstypes.NotificationChannelConfig{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &devopsguru.AddNotificationChannelInput{
		Config: cfg,
	}

	out, err := conn.AddNotificationChannel(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionCreating, ResNameNotificationChannel, "", err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionCreating, ResNameNotificationChannel, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state notificationChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findNotificationChannelByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionSetting, ResNameNotificationChannel, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.Id)

	resp.Diagnostics.Append(flex.Flatten(ctx, out.Config, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *notificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().DevOpsGuruClient(ctx)

	var state notificationChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &devopsguru.RemoveNotificationChannelInput{
		Id: state.ID.ValueStringPointer(),
	}

	_, err := conn.RemoveNotificationChannel(ctx, in)
	if err != nil {
		if errs.IsA[*retry.NotFoundError](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DevOpsGuru, create.ErrActionDeleting, ResNameNotificationChannel, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findNotificationChannelByID(ctx context.Context, conn *devopsguru.Client, id string) (*awstypes.NotificationChannel, error) {
	in := &devopsguru.ListNotificationChannelsInput{}

	paginator := devopsguru.NewListNotificationChannelsPaginator(conn, in)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, channel := range page.Channels {
			if aws.ToString(channel.Id) == id {
				return &channel, nil
			}
		}
	}

	return nil, &retry.NotFoundError{
		LastError:   errors.New("not found"),
		LastRequest: in,
	}
}

type notificationChannelResourceModel struct {
	framework.WithRegionModel
	Filters fwtypes.ListNestedObjectValueOf[filtersData] `tfsdk:"filters"`
	ID      types.String                                 `tfsdk:"id"`
	Sns     fwtypes.ListNestedObjectValueOf[snsData]     `tfsdk:"sns"`
}

type filtersData struct {
	MessageTypes fwtypes.SetValueOf[types.String] `tfsdk:"message_types"`
	Severities   fwtypes.SetValueOf[types.String] `tfsdk:"severities"`
}

type snsData struct {
	TopicARN fwtypes.ARN `tfsdk:"topic_arn"`
}
