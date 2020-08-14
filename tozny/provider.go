// Package tozny implements a Terraform Provider for automating provisioning of Tozny Services using Terraform.
package tozny

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a Terraform Provider for automating provisioning of Tozny Services using Terraform.
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"tozny_account": resourceAccount(),
		},
	}
}
