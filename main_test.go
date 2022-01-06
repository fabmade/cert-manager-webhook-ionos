package main

import (
	"github.com/fabmade/cert-manager-webhook-ionos/ionos"
	"github.com/jetstack/cert-manager/test/acme/dns"
	"os"
	"testing"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.
	//
	zone = "example.com." // todo remove it

	solver := ionos.NewMock("59351")

	fixture := dns.NewFixture(solver,
		dns.SetResolvedZone(zone),
		dns.SetDNSServer("127.0.0.1:59351"),
		dns.SetAllowAmbientCredentials(false),
		dns.SetManifestPath("testdata/ionos"),
		dns.SetUseAuthoritative(false),
	)
	fixture.RunConformance(t)
}
