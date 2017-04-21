package aws_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/application/aws"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BrokenEnvironmentValidator", func() {
	var (
		brokenEnvironmentValidator aws.BrokenEnvironmentValidator
		infrastructureManager      *fakes.InfrastructureManager
	)
	BeforeEach(func() {
		infrastructureManager = &fakes.InfrastructureManager{}

		brokenEnvironmentValidator = aws.NewBrokenEnvironmentValidator(infrastructureManager)
	})

	It("returns a helpful error message when the cloud formation doesn't exist but the bosh state does", func() {
		infrastructureManager.ExistsCall.Returns.Exists = false
		err := brokenEnvironmentValidator.Validate(storage.State{
			IAAS: "aws",
			AWS: storage.AWS{
				Region: "some-aws-region",
			},
			Stack: storage.Stack{
				Name: "some-stack-name",
			},
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		})

		Expect(err).To(MatchError("Found BOSH data in state directory, " +
			"but Cloud Formation stack \"some-stack-name\" cannot be found for region \"some-aws-region\" and given " +
			"AWS credentials. bbl cannot safely proceed. Open an issue on GitHub at " +
			"https://github.com/cloudfoundry/bosh-bootloader/issues/new if you need assistance."))

		Expect(infrastructureManager.ExistsCall.CallCount).To(Equal(1))
		Expect(infrastructureManager.ExistsCall.Receives.StackName).To(Equal("some-stack-name"))
	})

	It("returns no error if stack exists", func() {
		infrastructureManager.ExistsCall.Returns.Exists = true
		err := brokenEnvironmentValidator.Validate(storage.State{
			IAAS: "aws",
			AWS: storage.AWS{
				Region: "some-aws-region",
			},
			Stack: storage.Stack{
				Name: "some-stack-name",
			},
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns no error if the stack doesn't exist and BOSH state is empty", func() {
		infrastructureManager.ExistsCall.Returns.Exists = false
		err := brokenEnvironmentValidator.Validate(storage.State{
			IAAS: "aws",
			AWS: storage.AWS{
				Region: "some-aws-region",
			},
			Stack: storage.Stack{
				Name: "some-stack-name",
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when terraform was used to create infrastructure", func() {
		It("no ops", func() {
			infrastructureManager.ExistsCall.Returns.Exists = false
			err := brokenEnvironmentValidator.Validate(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				TFState: "some-tf-state",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(infrastructureManager.ExistsCall.CallCount).To(Equal(0))
		})
	})

	Context("failure cases", func() {
		It("returns an error when infrastructure manager exists fails", func() {
			infrastructureManager.ExistsCall.Returns.Error = errors.New("failed to check state")
			err := brokenEnvironmentValidator.Validate(storage.State{
				IAAS: "aws",
				AWS: storage.AWS{
					Region: "some-aws-region",
				},
				Stack: storage.Stack{
					Name: "some-stack-name",
				},
			})
			Expect(err).To(MatchError("failed to check state"))
		})
	})
})
