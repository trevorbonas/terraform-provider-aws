// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerObservabilityConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`observabilityconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ObservabilityConfigurationStatusActive)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppRunnerObservabilityConfiguration_traceConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityConfigurationConfig_traceConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "apprunner", regexache.MustCompile(fmt.Sprintf(`observabilityconfiguration/%s/1/.+`, rName))),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_name", rName),
					resource.TestCheckResourceAttr(resourceName, "observability_configuration_revision", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, string(types.ObservabilityConfigurationStatusActive)),
					resource.TestCheckResourceAttr(resourceName, "trace_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "trace_configuration.0.vendor", "AWSXRAY"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppRunnerObservabilityConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObservabilityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfapprunner.ResourceObservabilityConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppRunnerObservabilityConfiguration_Identity_ExistingResource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_apprunner_observability_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_12_0),
		},
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.AppRunnerServiceID),
		CheckDestroy: testAccCheckObservabilityConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.100.0",
					},
				},
				Config: testAccObservabilityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					tfstatecheck.ExpectNoIdentity(resourceName),
				},
			},
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "6.0.0",
					},
				},
				Config: testAccObservabilityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: knownvalue.Null(),
					}),
				},
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccObservabilityConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObservabilityConfigurationExists(ctx, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNRegexp("apprunner", regexache.MustCompile(`observabilityconfiguration/.+`)),
					}),
				},
			},
		},
	})
}

func testAccCheckObservabilityConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_observability_configuration" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

			_, err := tfapprunner.FindObservabilityConfigurationByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Runner Observability Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckObservabilityConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No App Runner Service ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppRunnerClient(ctx)

		_, err := tfapprunner.FindObservabilityConfigurationByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccObservabilityConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = %[1]q
}
`, rName)
}

func testAccObservabilityConfigurationConfig_traceConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_observability_configuration" "test" {
  observability_configuration_name = %[1]q
  trace_configuration {
    vendor = "AWSXRAY"
  }
}
`, rName)
}
