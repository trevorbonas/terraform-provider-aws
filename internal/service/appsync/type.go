// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appsync_type")
func ResourceType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTypeCreate,
		ReadWithoutTimeout:   resourceTypeRead,
		UpdateWithoutTimeout: resourceTypeUpdate,
		DeleteWithoutTimeout: resourceTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrFormat: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appsync.TypeDefinitionFormat_Values(), false),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID := d.Get("api_id").(string)

	params := &appsync.CreateTypeInput{
		ApiId:      aws.String(apiID),
		Definition: aws.String(d.Get("definition").(string)),
		Format:     aws.String(d.Get(names.AttrFormat).(string)),
	}

	out, err := conn.CreateTypeWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Appsync Type: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", apiID, aws.StringValue(out.Type.Format), aws.StringValue(out.Type.Name)))

	return append(diags, resourceTypeRead(ctx, d, meta)...)
}

func resourceTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	apiID, format, name, err := DecodeTypeID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Type %q: %s", d.Id(), err)
	}

	resp, err := FindTypeByThreePartKey(ctx, conn, apiID, format, name)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync Type (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Appsync Type %q: %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set(names.AttrARN, resp.Arn)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrFormat, resp.Format)
	d.Set("definition", resp.Definition)
	d.Set(names.AttrDescription, resp.Description)

	return diags
}

func resourceTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	params := &appsync.UpdateTypeInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		Format:     aws.String(d.Get(names.AttrFormat).(string)),
		TypeName:   aws.String(d.Get(names.AttrName).(string)),
		Definition: aws.String(d.Get("definition").(string)),
	}

	_, err := conn.UpdateTypeWithContext(ctx, params)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Appsync Type %q: %s", d.Id(), err)
	}

	return append(diags, resourceTypeRead(ctx, d, meta)...)
}

func resourceTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppSyncConn(ctx)

	input := &appsync.DeleteTypeInput{
		ApiId:    aws.String(d.Get("api_id").(string)),
		TypeName: aws.String(d.Get(names.AttrName).(string)),
	}
	_, err := conn.DeleteTypeWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Appsync Type: %s", err)
	}

	return diags
}

func DecodeTypeID(id string) (string, string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("Unexpected format of ID (%q), expected API-ID:FORMAT:TYPE-NAME", id)
	}
	return parts[0], parts[1], parts[2], nil
}
