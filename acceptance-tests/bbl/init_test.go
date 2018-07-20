package acceptance_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bbl")
}

var (
	pathToBBL             string
	bblPlanTimeout        time.Duration
	bblDownTimeout        time.Duration
	bblUpTimeout          time.Duration
	bblRotateTimeout      time.Duration
	bblLatestErrorTimeout time.Duration
	bblLeftoversTimeout   time.Duration
)

var _ = BeforeSuite(func() {
	var err error

	pathToBBL, err = gexec.Build("github.com/cloudfoundry/bosh-bootloader/bbl")
	Expect(err).NotTo(HaveOccurred())

	bblPlanTimeout = getTimeout("BBL_PLAN_TIMEOUT", 5*time.Minute)
	bblDownTimeout = getTimeout("BBL_DOWN_TIMEOUT", 10*time.Minute)
	bblUpTimeout = getTimeout("BBL_UP_TIMEOUT", 40*time.Minute)
	bblRotateTimeout = getTimeout("BBL_ROTATE_TIMEOUT", 40*time.Minute)
	bblLatestErrorTimeout = getTimeout("BBL_LATEST_ERROR_TIMEOUT", 10*time.Second)
	bblLeftoversTimeout = getTimeout("BBL_LEFTOVERS_TIMEOUT", 10*time.Minute)
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func getTimeout(envVar string, defaultTimeout time.Duration) time.Duration {
	envVal := os.Getenv(envVar)
	if envVal == "" {
		return defaultTimeout
	}
	timeout, err := time.ParseDuration(envVal)
	Expect(err).NotTo(HaveOccurred())
	return timeout
}
