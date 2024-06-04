// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/timestreaminfluxdb/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_timestreaminfluxdb_db_instance", name="Db Instance")
// @Tags(identifierAttribute="arn")
func newResourceDbInstance(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDbInstance{}

	// TIP: ==== CONFIGURABLE TIMEOUTS ====
	// Users can configure timeout lengths but you need to use the times they
	// provide. Access the timeout they configure (or the defaults) using,
	// e.g., r.CreateTimeout(ctx, plan.Timeouts) (see below). The times here are
	// the defaults if they don't configure timeouts.
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	// If not provided, CreateDbInstance will use the below default values
	// for bucket and organization. These values need to be set in Terraform
	// because GetDbInstance won't return them.
	DefaultBucketValue       = "bucket"
	DefaultOrganizationValue = "organization"
	ResNameDbInstance        = "Db Instance"
)

type resourceDbInstance struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDbInstance) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_timestreaminfluxdb_db_instance"
}

// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
//   - If a user can provide a value ("configure a value") for an
//     attribute (e.g., instances = 5), we call the attribute an
//     "argument."
//   - You change the way users interact with attributes using:
//   - Required
//   - Optional
//   - Computed
//   - There are only four valid combinations:
//
// 1. Required only - the user must provide a value
// Required: true,
//
//  2. Optional only - the user can configure or omit a value; do not
//     use Default or DefaultFunc
//
// Optional: true,
//
//  3. Computed only - the provider can provide a value but the user
//     cannot, i.e., read-only
//
// Computed: true,
//
//  4. Optional AND Computed - the provider or user can provide a value;
//     use this combination if you are using Default
//
// Optional: true,
// Computed: true,
//
// You will typically find arguments in the input struct
// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
// they are only in the input struct (e.g., ModifyDBInstanceInput) for
// the modify operation.
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
func (r *resourceDbInstance) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"allocated_storage": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(20),
					int64validator.AtMost(16384),
				},
				Description: `The amount of storage to allocate for your DB storage type in GiB (gibibytes). 
					This argument is required. This argument has a minimum value of 
					20 and a maximum value of 16384`,
			},
			"arn": framework.ARNAttributeComputedOnly(),
			"availability_zone": schema.StringAttribute{
				Computed: true,
			},
			"bucket": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(2),
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(
						// Taken from the model for TimestreamInfluxDB in AWS SDK Go V2
						// https://github.com/aws/aws-sdk-go-v2/blob/8209abb7fa1aeb513228b4d8c1a459aeb6209d4d/codegen/sdk-codegen/aws-models/timestream-influxdb.json#L768
						regexache.MustCompile("^[^_][^\"]*$"),
						"",
					),
				},
				Description: `The name of the initial InfluxDB bucket. All InfluxDB data is stored in a bucket. 
					A bucket combines the concept of a database and a retention period (the duration of time 
					that each data point persists). A bucket belongs to an organization. This argument is optional. 
					If not provided, defaults to "bucket"`,
			},
			"db_instance_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"db.influx.medium",
						"db.influx.large",
						"db.influx.xlarge",
						"db.influx.2xlarge",
						"db.influx.4xlarge",
						"db.influx.8xlarge",
						"db.influx.12xlarge",
						"db.influx.16xlarge",
					),
				},
				Description: `The Timestream for InfluxDB DB instance type to run InfluxDB on. 
					This argument is required. Possible values: 
					"db.influx.medium", 
					"db.influx.large", 
					"db.influx.xlarge", 
					"db.influx.2xlarge", 
					"db.influx.4xlarge", 
					"db.influx.8xlarge", 
					"db.influx.12xlarge", 
					"db.influx.16xlarge"`,
			},
			"db_parameter_group_identifier": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(
						// Taken from the model for TimestreamInfluxDB in AWS SDK Go V2
						// https://github.com/aws/aws-sdk-go-v2/blob/8209abb7fa1aeb513228b4d8c1a459aeb6209d4d/codegen/sdk-codegen/aws-models/timestream-influxdb.json#L1390
						regexache.MustCompile("^[a-zA-Z0-9]+$"),
						"",
					),
				},
			},
			"db_storage_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"InfluxIOIncludedT1",
						"InfluxIOIncludedT2",
						"InfluxIOIncludedT3",
					),
				},
				Description: `The Timestream for InfluxDB DB storage type to read and write InfluxDB data. 
					You can choose between 3 different types of provisioned Influx IOPS included storage according 
					to your workloads requirements: Influx IO Included 3000 IOPS, Influx IO Included 12000 IOPS, 
					Influx IO Included 16000 IOPS. This argument is optional. Possible values: 
					"InfluxIOIncludedT1", 
					"InfluxIOIncludedT2", 
					"InfluxIOIncludedT3". 
					If not provided, defaults to "InfluxIOIncludedT1"`,
			},
			"deployment_type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"SINGLE_AZ",
						"WITH_MULTIAZ_STANDBY",
					),
				},
				Description: `Specifies whether the DB instance will be deployed as a standalone instance or 
					with a Multi-AZ standby for high availability. This argument is optional. 
					Possible values: 
					"SINGLE_AZ", 
					"WITH_MULTIAZ_STANDBY". 
					If not provided, defaults to "SINGLE_AZ"`,
			},
			"endpoint": schema.StringAttribute{
				Computed: true,
			},
			"id":                                framework.IDAttribute(),
			"influx_auth_parameters_secret_arn": framework.ARNAttributeComputedOnly(),
			"name": schema.StringAttribute{
				Required: true,
				// TIP: ==== PLAN MODIFIERS ====
				// Plan modifiers were introduced with Plugin-Framework to provide a mechanism
				// for adjusting planned changes prior to apply. The planmodifier subpackage
				// provides built-in modifiers for many common use cases such as
				// requiring replacement on a value change ("ForceNew: true" in Plugin-SDK
				// resources).
				//
				// See more:
				// https://developer.hashicorp.com/terraform/plugin/framework/resources/plan-modification
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(3),
					stringvalidator.LengthAtMost(40),
					stringvalidator.RegexMatches(
						// Taken from the model for TimestreamInfluxDB in AWS SDK Go V2
						// https://github.com/aws/aws-sdk-go-v2/blob/8209abb7fa1aeb513228b4d8c1a459aeb6209d4d/codegen/sdk-codegen/aws-models/timestream-influxdb.json#L1215
						regexache.MustCompile("^[a-zA-z][a-zA-Z0-9]*(-[a-zA-Z0-9]+)*$"),
						"",
					),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"organization": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(8),
					stringvalidator.LengthAtMost(64),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9]+$"), ""),
				},
			},
			"publicly_accessible": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"secondary_availability_zone": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile("^[a-zA-Z]([a-zA-Z0-9]*(-[a-zA-Z0-9]+)*)?$"),
						`Must start with a letter and can't end with a hyphen or contain two 
						consecutive hyphens`,
					),
				},
				Description: `The username of the initial admin user created in InfluxDB. 
					Must start with a letter and can't end with a hyphen or contain two 
					consecutive hyphens. For example, my-user1. This username will allow 
					you to access the InfluxDB UI to perform various administrative tasks 
					and also use the InfluxDB CLI to create an operator token. These 
					attributes will be stored in a Secret created in Amazon Secrets 
					Manager in your account`,
			},
			"vpc_security_group_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(5),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtMost(64),
						stringvalidator.RegexMatches(regexache.MustCompile("^sg-[a-z0-9]+$"), ""),
					),
				},
			},
			"vpc_subnet_ids": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.SizeAtMost(3),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtMost(64),
						stringvalidator.RegexMatches(regexache.MustCompile("^subnet-[a-z0-9]+$"), ""),
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"log_delivery_configuration": schema.ListNestedBlock{
				// TIP: ==== LIST VALIDATORS ====
				// List and set validators take the place of MaxItems and MinItems in
				// Plugin-Framework based resources. Use listvalidator.SizeAtLeast(1) to
				// make a nested object required. Similar to Plugin-SDK, complex objects
				// can be represented as lists or sets with listvalidator.SizeAtMost(1).
				//
				// For a complete mapping of Plugin-SDK to Plugin-Framework schema fields,
				// see:
				// https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/blocks
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"s3_configuration": schema.SingleNestedBlock{
							Attributes: map[string]schema.Attribute{
								"bucket_name": schema.StringAttribute{
									Optional: true,
									Validators: []validator.String{
										stringvalidator.LengthAtLeast(3),
										stringvalidator.LengthAtMost(63),
										stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-z]+[0-9a-z\\.\\-]*[0-9a-z]+$"), ""),
									},
								},
								"enabled": schema.BoolAttribute{
									Optional: true,
								},
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDbInstance) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan
	// 3. Populate a create input structure
	// 4. Call the AWS create/put function
	// 5. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work, as well as any computed
	//    only attributes.
	// 6. Use a waiter to wait for create to complete
	// 7. Save the request plan to response state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan resourceDbInstanceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a create input structure
	in := &timestreaminfluxdb.CreateDbInstanceInput{
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		AllocatedStorage:    aws.Int32(int32(plan.AllocatedStorage.ValueInt64())),
		DbInstanceType:      awstypes.DbInstanceType(plan.DBInstanceType.ValueString()),
		Name:                aws.String(plan.Name.ValueString()),
		Password:            aws.String(plan.Password.ValueString()),
		VpcSecurityGroupIds: flex.ExpandFrameworkStringValueSet(ctx, plan.VPCSecurityGroupIDs),
		VpcSubnetIds:        flex.ExpandFrameworkStringValueSet(ctx, plan.VPCSubnetIDs),
		Tags:                getTagsIn(ctx),
	}
	if plan.Bucket.IsNull() || plan.Bucket.IsUnknown() {
		plan.Bucket = types.StringValue(DefaultBucketValue)
	}
	in.Bucket = aws.String(plan.Bucket.ValueString())
	if !plan.DBParameterGroupIdentifier.IsNull() {
		in.DbParameterGroupIdentifier = aws.String(plan.DBParameterGroupIdentifier.ValueString())
	}
	if !plan.DBStorageType.IsNull() {
		in.DbStorageType = awstypes.DbStorageType(plan.DBStorageType.ValueString())
	}
	if !plan.DeploymentType.IsNull() {
		in.DeploymentType = awstypes.DeploymentType(plan.DeploymentType.ValueString())
	}
	// TIP: Use an expander to assign a complex argument. The elements must be
	// deserialized into the appropriate struct before being passed to the expander.
	if !plan.LogDeliveryConfiguration.IsNull() {
		var tfList []logDeliveryConfigurationData
		resp.Diagnostics.Append(plan.LogDeliveryConfiguration.ElementsAs(ctx, &tfList, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		in.LogDeliveryConfiguration = expandLogDeliveryConfiguration(tfList)
	}
	if plan.Organization.IsNull() || plan.Organization.IsUnknown() {
		plan.Organization = types.StringValue(DefaultOrganizationValue)
	}
	in.Organization = aws.String(plan.Organization.ValueString())
	if !plan.PubliclyAccessible.IsNull() {
		in.PubliclyAccessible = aws.Bool(plan.PubliclyAccessible.ValueBool())
	}
	if !plan.Username.IsNull() {
		in.Username = aws.String(plan.Username.ValueString())
	}

	// TIP: -- 4. Call the AWS create function
	out, err := conn.CreateDbInstance(ctx, in)
	if err != nil {
		// TIP: Since ID has not been set yet, you cannot use plan.ID.String()
		// in error messages at this point.
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionCreating, ResNameDbInstance, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionCreating, ResNameDbInstance, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// TIP: -- 5. Using the output from the create function, set the minimum attributes
	plan.ARN = flex.StringToFramework(ctx, out.Arn)
	plan.ID = flex.StringToFramework(ctx, out.Id)
	plan.AvailabilityZone = flex.StringToFramework(ctx, out.AvailabilityZone)
	plan.DBParameterGroupIdentifier = flex.StringToFramework(ctx, out.DbParameterGroupIdentifier)
	logDeliveryConfiguration, d := flattenLogDeliveryConfiguration(ctx, out.LogDeliveryConfiguration)
	resp.Diagnostics.Append(d...)
	plan.LogDeliveryConfiguration = logDeliveryConfiguration
	plan.InfluxAuthParametersSecretARN = flex.StringToFramework(ctx, out.InfluxAuthParametersSecretArn)
	plan.DBParameterGroupIdentifier = flex.StringToFramework(ctx, out.DbParameterGroupIdentifier)
	plan.DBStorageType = flex.StringToFramework(ctx, (*string)(&out.DbStorageType))
	plan.DeploymentType = flex.StringToFramework(ctx, (*string)(&out.DeploymentType))
	plan.PubliclyAccessible = flex.BoolToFramework(ctx, out.PubliclyAccessible)

	// TIP: -- 6. Use a waiter to wait for create to complete
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDbInstanceCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForCreation, ResNameDbInstance, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	readOut, err := findDbInstanceByID(ctx, conn, plan.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDbInstance, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Attributes only set after resource is finished creating
	plan.Endpoint = flex.StringToFramework(ctx, out.Endpoint)
	plan.Status = flex.StringToFramework(ctx, (*string)(&readOut.Status))
	plan.SecondaryAvailabilityZone = flex.StringToFramework(ctx, readOut.SecondaryAvailabilityZone)

	// TIP: -- 7. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDbInstance) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Get the resource from AWS
	// 4. Remove resource from state if it is not found
	// 5. Set the arguments and attributes
	// 6. Set the state

	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceDbInstanceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findDbInstanceByID(ctx, conn, state.ID.ValueString())
	// TIP: -- 4. Remove resource from state if it is not found
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDbInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.StringAttribute, schema.BoolAttribute,
	// schema.Int64Attribute, and schema.Float64Attribue), simply setting the
	// appropriate data struct field is sufficient. The flex package implements
	// helpers for converting between Go and Plugin-Framework types seamlessly. No
	// error or nil checking is necessary.
	//
	// However, there are some situations where more handling is needed such as
	// complex data types (e.g., schema.ListAttribute, schema.SetAttribute). In
	// these cases the flatten function may have a diagnostics return value, which
	// should be appended to resp.Diagnostics.
	state.ARN = flex.StringToFramework(ctx, out.Arn)
	state.AllocatedStorage = flex.Int32ToFramework(ctx, out.AllocatedStorage)
	state.AvailabilityZone = flex.StringToFramework(ctx, out.AvailabilityZone)
	state.DBInstanceType = flex.StringToFramework(ctx, (*string)(&out.DbInstanceType))
	state.DBParameterGroupIdentifier = flex.StringToFramework(ctx, out.DbParameterGroupIdentifier)
	state.DBStorageType = flex.StringToFramework(ctx, (*string)(&out.DbStorageType))
	state.DeploymentType = flex.StringToFramework(ctx, (*string)(&out.DeploymentType))
	state.Endpoint = flex.StringToFramework(ctx, out.Endpoint)
	state.ID = flex.StringToFramework(ctx, out.Id)
	state.InfluxAuthParametersSecretARN = flex.StringToFramework(ctx, out.InfluxAuthParametersSecretArn)
	logDeliveryConfiguration, d := flattenLogDeliveryConfiguration(ctx, out.LogDeliveryConfiguration)
	resp.Diagnostics.Append(d...)
	state.LogDeliveryConfiguration = logDeliveryConfiguration
	state.Name = flex.StringToFramework(ctx, out.Name)
	state.PubliclyAccessible = flex.BoolToFramework(ctx, out.PubliclyAccessible)
	state.SecondaryAvailabilityZone = flex.StringToFramework(ctx, out.SecondaryAvailabilityZone)
	state.Status = flex.StringToFramework(ctx, (*string)(&out.Status))
	state.VPCSecurityGroupIDs = flex.FlattenFrameworkStringValueSet[string](ctx, out.VpcSecurityGroupIds)
	state.VPCSubnetIDs = flex.FlattenFrameworkStringValueSet[string](ctx, out.VpcSubnetIds)

	tags, err := listTags(ctx, conn, state.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionSetting, ResNameDbInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	setTagsOut(ctx, Tags(tags))

	// TIP: -- 6. Set the state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDbInstance) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have RequiresReplace() plan modifiers
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the resource will not have an update method
	// defined. In the case of c., Update and Create can be refactored to call
	// the same underlying function.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the plan and state
	// 3. Populate a modify input structure and check for changes
	// 4. Call the AWS modify/update function
	// 5. Use a waiter to wait for update to complete
	// 6. Save the request plan to response state
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	// TIP: -- 2. Fetch the plan
	var plan, state resourceDbInstanceData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a modify input structure and check for changes
	if !plan.Name.Equal(state.Name) ||
		!plan.LogDeliveryConfiguration.Equal(state.LogDeliveryConfiguration) ||
		!plan.DBInstanceType.Equal(state.DBInstanceType) {

		in := &timestreaminfluxdb.UpdateDbInstanceInput{
			// TIP: Mandatory or fields that will always be present can be set when
			// you create the Input structure. (Replace these with real fields.)
			Identifier: aws.String(plan.ID.ValueString()),
		}

		/*if !plan.Description.IsNull() {
			// TIP: Optional fields should be set based on whether or not they are
			// used.
			in.Description = aws.String(plan.Description.ValueString())
		}
		if !plan.ComplexArgument.IsNull() {
			// TIP: Use an expander to assign a complex argument. The elements must be
			// deserialized into the appropriate struct before being passed to the expander.
			var tfList []complexArgumentData
			resp.Diagnostics.Append(plan.ComplexArgument.ElementsAs(ctx, &tfList, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			in.ComplexArgument = expandComplexArgument(tfList)
		}*/

		// TIP: -- 4. Call the AWS modify/update function
		out, err := conn.UpdateDbInstance(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionUpdating, ResNameDbInstance, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Id == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionUpdating, ResNameDbInstance, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		// TIP: Using the output from the update function, re-set any computed attributes
		plan.ARN = flex.StringToFramework(ctx, out.Arn)
		plan.ID = flex.StringToFramework(ctx, out.Id)
	}

	// TIP: -- 5. Use a waiter to wait for update to complete
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitDbInstanceUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForUpdate, ResNameDbInstance, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 6. Save the request plan to response state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDbInstance) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Fetch the state
	// 3. Populate a delete input structure
	// 4. Call the AWS delete function
	// 5. Use a waiter to wait for delete to complete
	// TIP: -- 1. Get a client connection to the relevant service
	conn := r.Meta().TimestreamInfluxDBClient(ctx)

	// TIP: -- 2. Fetch the state
	var state resourceDbInstanceData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TIP: -- 3. Populate a delete input structure
	in := &timestreaminfluxdb.DeleteDbInstanceInput{
		Identifier: aws.String(state.ID.ValueString()),
	}

	// TIP: -- 4. Call the AWS delete function
	_, err := conn.DeleteDbInstance(ctx, in)
	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionDeleting, ResNameDbInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDbInstanceDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.TimestreamInfluxDB, create.ErrActionWaitingForDeletion, ResNameDbInstance, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

// TIP: ==== TERRAFORM IMPORTING ====
// If Read can get all the information it needs from the Identifier
// (i.e., path.Root("id")), you can use the PassthroughID importer. Otherwise,
// you'll need a custom import function.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/resources/import
func (r *resourceDbInstance) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *resourceDbInstance) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., awstypes.StatusInProgress).
/*const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)*/

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitDbInstanceCreated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.CreateDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.StatusCreating), string(awstypes.StatusUpdating), string(awstypes.StatusModifying)},
		Target:                    []string{string(awstypes.StatusAvailable)},
		Refresh:                   statusDbInstance(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.CreateDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitDbInstanceUpdated(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.UpdateDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.StatusModifying), string(awstypes.StatusUpdating)},
		Target:                    []string{string(awstypes.StatusAvailable)},
		Refresh:                   statusDbInstance(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.UpdateDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitDbInstanceDeleted(ctx context.Context, conn *timestreaminfluxdb.Client, id string, timeout time.Duration) (*timestreaminfluxdb.DeleteDbInstanceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.StatusDeleting), string(awstypes.StatusModifying), string(awstypes.StatusUpdating), string(awstypes.StatusAvailable)},
		Target:  []string{},
		Refresh: statusDbInstance(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*timestreaminfluxdb.DeleteDbInstanceOutput); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusDbInstance(ctx context.Context, conn *timestreaminfluxdb.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDbInstanceByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findDbInstanceByID(ctx context.Context, conn *timestreaminfluxdb.Client, id string) (*timestreaminfluxdb.GetDbInstanceOutput, error) {
	in := &timestreaminfluxdb.GetDbInstanceInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetDbInstance(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenLogDeliveryConfiguration(ctx context.Context, apiObject *awstypes.LogDeliveryConfiguration) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: logDeliveryConfigrationAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}
	s3Configuration, d := flattenS3Configuration(ctx, apiObject.S3Configuration)
	obj := map[string]attr.Value{
		"s3_configuration": s3Configuration,
	}
	objVal, d := types.ObjectValue(logDeliveryConfigrationAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func flattenS3Configuration(ctx context.Context, apiObject *awstypes.S3Configuration) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: s3ConfigurationAttrTypes}

	if apiObject == nil {
		return types.ObjectNull(elemType.AttrTypes), diags
	}

	obj := map[string]attr.Value{
		"bucket_name": flex.StringValueToFramework(ctx, *apiObject.BucketName),
		"enabled":     flex.BoolToFramework(ctx, *&apiObject.Enabled),
	}
	objVal, d := types.ObjectValue(s3ConfigurationAttrTypes, obj)
	diags.Append(d...)

	return objVal, diags
}

// TIP: ==== FLEX ====
// Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return the equivalent Plugin-Framework
// type. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
/*func flattenComplexArgument(ctx context.Context, apiObject *awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

	if apiObject == nil {
		return types.ListNull(elemType), diags
	}

	obj := map[string]attr.Value{
		"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
		"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
	}
	objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}*/

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
/*func flattenComplexArguments(ctx context.Context, apiObjects []*awstypes.ComplexArgument) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: complexArgumentAttrTypes}

	if len(apiObjects) == 0 {
		return types.ListNull(elemType), diags
	}

	elems := []attr.Value{}
	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		obj := map[string]attr.Value{
			"nested_required": flex.StringValueToFramework(ctx, apiObject.NestedRequired),
			"nested_optional": flex.StringValueToFramework(ctx, apiObject.NestedOptional),
		}
		objVal, d := types.ObjectValue(complexArgumentAttrTypes, obj)
		diags.Append(d...)

		elems = append(elems, objVal)
	}

	listVal, d := types.ListValue(elemType, elems)
	diags.Append(d...)

	return listVal, diags
}*/

func expandLogDeliveryConfiguration(tfList []logDeliveryConfigurationData) *awstypes.LogDeliveryConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.LogDeliveryConfiguration{
		S3Configuration: expandS3Configuration(tfObj.S3Configuration),
	}
	return apiObject
}

func expandS3Configuration(tfObj s3ConfigurationData) *awstypes.S3Configuration {
	apiObject := &awstypes.S3Configuration{
		BucketName: aws.String(tfObj.BucketName.ValueString()),
		Enabled:    aws.Bool(tfObj.Enabled.ValueBool()),
	}
	return apiObject
}

// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/
/*func expandComplexArgument(tfList []complexArgumentData) *awstypes.ComplexArgument {
	if len(tfList) == 0 {
		return nil
	}

	tfObj := tfList[0]
	apiObject := &awstypes.ComplexArgument{
		NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
	}
	if !tfObj.NestedOptional.IsNull() {
		apiObject.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
	}

	return apiObject
}*/

// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
/*func expandComplexArguments(tfList []complexArgumentData) []*timestreaminfluxdb.ComplexArgument {
	// TIP: The AWS API can be picky about whether you send a nil or zero-
	// length for an argument that should be cleared. For example, in some
	// cases, if you send a nil value, the AWS API interprets that as "make no
	// changes" when what you want to say is "remove everything." Sometimes
	// using a zero-length list will cause an error.
	//
	// As a result, here are two options. Usually, option 1, nil, will work as
	// expected, clearing the field. But, test going from something to nothing
	// to make sure it works. If not, try the second option.
	// TIP: Option 1: Returning nil for zero-length list
    if len(tfList) == 0 {
        return nil
    }
    var apiObject []*awstypes.ComplexArgument
	// TIP: Option 2: Return zero-length list for zero-length list. If option 1 does
	// not work, after testing going from something to nothing (if that is
	// possible), uncomment out the next line and remove option 1.
	//
	// apiObject := make([]*timestreaminfluxdb.ComplexArgument, 0)

	for _, tfObj := range tfList {
		item := &timestreaminfluxdb.ComplexArgument{
			NestedRequired: aws.String(tfObj.NestedRequired.ValueString()),
		}
		if !tfObj.NestedOptional.IsNull() {
			item.NestedOptional = aws.String(tfObj.NestedOptional.ValueString())
		}

		apiObject = append(apiObject, item)
	}

	return apiObject
}*/

// TIP: ==== DATA STRUCTURES ====
// With Terraform Plugin-Framework configurations are deserialized into
// Go types, providing type safety without the need for type assertions.
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
//
// Nested objects are represented in their own data struct. These will
// also have a corresponding attribute type mapping for use inside flex
// functions.
//
// See more:
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/accessing-values
type resourceDbInstanceData struct {
	AllocatedStorage              types.Int64    `tfsdk:"allocated_storage"`
	ARN                           types.String   `tfsdk:"arn"`
	AvailabilityZone              types.String   `tfsdk:"availability_zone"`
	Bucket                        types.String   `tfsdk:"bucket"`
	DBInstanceType                types.String   `tfsdk:"db_instance_type"`
	DBParameterGroupIdentifier    types.String   `tfsdk:"db_parameter_group_identifier"`
	DBStorageType                 types.String   `tfsdk:"db_storage_type"`
	DeploymentType                types.String   `tfsdk:"deployment_type"`
	Endpoint                      types.String   `tfsdk:"endpoint"`
	ID                            types.String   `tfsdk:"id"`
	InfluxAuthParametersSecretARN types.String   `tfsdk:"influx_auth_parameters_secret_arn"`
	LogDeliveryConfiguration      types.List     `tfsdk:"log_delivery_configuration"`
	Name                          types.String   `tfsdk:"name"`
	Organization                  types.String   `tfsdk:"organization"`
	Password                      types.String   `tfsdk:"password"`
	PubliclyAccessible            types.Bool     `tfsdk:"publicly_accessible"`
	SecondaryAvailabilityZone     types.String   `tfsdk:"secondary_availability_zone"`
	Status                        types.String   `tfsdk:"status"`
	Tags                          types.Map      `tfsdk:"tags"`
	TagsAll                       types.Map      `tfsdk:"tags_all"`
	Timeouts                      timeouts.Value `tfsdk:"timeouts"`
	Username                      types.String   `tfsdk:"username"`
	VPCSecurityGroupIDs           types.Set      `tfsdk:"vpc_security_group_ids"`
	VPCSubnetIDs                  types.Set      `tfsdk:"vpc_subnet_ids"`
}

type logDeliveryConfigurationData struct {
	S3Configuration s3ConfigurationData `tfsdk:"s3_configuration"`
}

type s3ConfigurationData struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Enabled    types.Bool   `tfsdk:"enabled"`
}

var logDeliveryConfigrationAttrTypes = map[string]attr.Type{
	"s3_configuration": types.ObjectType{AttrTypes: s3ConfigurationAttrTypes},
}

var s3ConfigurationAttrTypes = map[string]attr.Type{
	"bucket_name": types.StringType,
	"enabled":     types.BoolType,
}
