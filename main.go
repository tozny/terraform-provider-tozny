package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/tozny/terraform-provider-tozny/tozny"
)

func main() {
	// Parse command line debug flag if present
	var debugMode bool
	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()
	// If debug mode specified run a debug server in the background
	// https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-debuggable-provider-binaries
	if debugMode {
		err := plugin.Debug(context.Background(), "terraform.tozny.com/tozny/tozny",
			&plugin.ServeOpts{
				ProviderFunc: tozny.Provider,
			})
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		// Otherwise run just the provider
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: tozny.Provider})
	}
}
