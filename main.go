package main

import (
	"github.com/fabmade/cert-manager-webhook-ionos/ionos"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"os"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {

	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName, ionos.New())
}
