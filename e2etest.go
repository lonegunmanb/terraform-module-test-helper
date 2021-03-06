package terraform_module_test_helper

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

type TerraformOutput = map[string]interface{}

func RunE2ETest(t *testing.T, moduleRootPath, exampleRelativePath string, option terraform.Options, assertion func(*testing.T, TerraformOutput)) {
	option = retryableOptions(t, option)
	tmpDir := test_structure.CopyTerraformFolderToTemp(t, moduleRootPath, exampleRelativePath)
	err := renderOverrideFile(tmpDir)
	if err != nil {
		t.Errorf(err.Error())
	}
	option.TerraformDir = tmpDir
	defer terraform.Destroy(t, &option)
	terraform.InitAndApply(t, &option)
	err = initAndPlanAndIdempotentAtEasyMode(t, option)
	if err != nil {
		t.Errorf(err.Error())
	}
	if assertion != nil {
		noLoggerOption, err := option.Clone()
		if err != nil {
			t.Errorf(err.Error())
		}
		// default logger might leak sensitive data
		noLoggerOption.Logger = logger.Discard
		assertion(t, terraform.OutputAll(t, noLoggerOption))
	}
}

func retryableOptions(t *testing.T, options terraform.Options) terraform.Options {
	result := terraform.WithDefaultRetryableErrors(t, &options)
	result.RetryableTerraformErrors[".*Please try again.*"] = "Service side suggest retry."
	return *result
}