package cql

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

func TestCassandraMigrate(t *testing.T) {
	RegisterFailHandler(Fail)
	if os.Getenv("TEAMCITY") == "true" {
		RunSpecsWithCustomReporters(t, "Cassandra CQL", []Reporter{reporters.NewTeamCityReporter(os.Stdout)})
	} else {
		RunSpecs(t, "Cassandra CQL")
	}
}
