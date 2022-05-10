package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const retryMsg = "Retry the Terraform operation. If the error still occurs or happens regularly, please contact the provider developer with hardware and operating system information.\n\n"

func randomReadError(errMsg string) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.AddError(
		"Random Read Error",
		"While attempting to generate a random value for this resource, a read error was generated.\n\n"+
			retryMsg+
			fmt.Sprintf("Original Error: %s", errMsg),
	)

	return diags
}
