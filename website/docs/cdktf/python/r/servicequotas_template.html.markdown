---
subcategory: "Service Quotas"
layout: "aws"
page_title: "AWS: aws_servicequotas_template"
description: |-
  Terraform resource for managing an AWS Service Quotas Template.
---

<!-- Please do not edit this file, it is generated. -->
# Resource: aws_servicequotas_template

Terraform resource for managing an AWS Service Quotas Template.

-> Only the management account of an organization can alter Service Quota templates, and this must be done from the `us-east-1` region.

## Example Usage

### Basic Usage

```python
# DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
from constructs import Construct
from cdktf import Token, TerraformStack
#
# Provider bindings are generated by running `cdktf get`.
# See https://cdk.tf/provider-generation for more details.
#
from imports.aws.servicequotas_template import ServicequotasTemplate
class MyConvertedCode(TerraformStack):
    def __init__(self, scope, name):
        super().__init__(scope, name)
        ServicequotasTemplate(self, "example",
            quota_code="L-2ACBD22F",
            region="us-east-1",
            service_code="lambda",
            value=Token.as_number("80")
        )
```

## Argument Reference

The following arguments are required:

* `region` - (Required) AWS Region to which the template applies.
* `quota_code` - (Required) Quota identifier. To find the quota code for a specific quota, use the [aws_servicequotas_service_quota](../d/servicequotas_service_quota.html.markdown) data source.
* `service_code` - (Required) Service identifier. To find the service code value for an AWS service, use the [aws_servicequotas_service](../d/servicequotas_service.html.markdown) data source.
* `value` - (Required) The new, increased value for the quota.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `global_quota` - Indicates whether the quota is global.
* `id` - Unique identifier for the resource, which is a comma-delimited string separating `region`, `quota_code`, and `service_code`.
* `quota_name` - Quota name.
* `service_name` - Service name.
* `unit` - Unit of measurement.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Service Quotas Template using the `id`. For example:

```python
# DO NOT EDIT. Code generated by 'cdktf convert' - Please report bugs at https://cdk.tf/bug
from constructs import Construct
from cdktf import TerraformStack
#
# Provider bindings are generated by running `cdktf get`.
# See https://cdk.tf/provider-generation for more details.
#
from imports.aws.servicequotas_template import ServicequotasTemplate
class MyConvertedCode(TerraformStack):
    def __init__(self, scope, name):
        super().__init__(scope, name)
        ServicequotasTemplate.generate_config_for_import(self, "example", "us-east-1,L-2ACBD22F,lambda")
```

Using `terraform import`, import Service Quotas Template using the `id`. For example:

```console
% terraform import aws_servicequotas_template.example us-east-1,L-2ACBD22F,lambda
```

<!-- cache-key: cdktf-0.20.8 input-0cf03b3edb756335c791cb160039a0bb5c8f09933443b5face04c802ab7c11c5 -->