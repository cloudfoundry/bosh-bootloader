package terraform

import (
	"bytes"
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/coreos/go-semver/semver"
)

type Manager struct {
	executor              executor
	templateGenerator     TemplateGenerator
	inputGenerator        InputGenerator
	outputGenerator       OutputGenerator
	terraformOutputBuffer *bytes.Buffer
	logger                logger
	stackMigrator         stackMigrator
}

type executor interface {
	Version() (string, error)
	Destroy(inputs map[string]string, terraformTemplate, tfState string) (string, error)
	Apply(inputs map[string]string, terraformTemplate, tfState string) (string, error)
}

//go:generate counterfeiter -o ./fakes/stack_migrator.go --fake-name StackMigrator . stackMigrator
type stackMigrator interface {
	Migrate(state storage.State) (storage.State, error)
}

type InputGenerator interface {
	Generate(storage.State) (map[string]string, error)
}

type OutputGenerator interface {
	Generate(tfState string) (map[string]interface{}, error)
}

type TemplateGenerator interface {
	Generate(storage.State) string
}

type logger interface {
	Step(string, ...interface{})
}

type NewManagerArgs struct {
	Executor              executor
	TemplateGenerator     TemplateGenerator
	InputGenerator        InputGenerator
	OutputGenerator       OutputGenerator
	TerraformOutputBuffer *bytes.Buffer
	Logger                logger
	StackMigrator         stackMigrator
}

func NewManager(args NewManagerArgs) Manager {
	return Manager{
		executor:              args.Executor,
		templateGenerator:     args.TemplateGenerator,
		inputGenerator:        args.InputGenerator,
		outputGenerator:       args.OutputGenerator,
		terraformOutputBuffer: args.TerraformOutputBuffer,
		logger:                args.Logger,
		stackMigrator:         args.StackMigrator,
	}
}

func (m Manager) Version() (string, error) {
	return m.executor.Version()
}

func (m Manager) ValidateVersion() error {
	version, err := m.executor.Version()
	if err != nil {
		return err
	}

	currentVersion, err := semver.NewVersion(version)
	if err != nil {
		return err
	}

	// This shouldn't fail, so there is no test for capturing the error.
	minimumVersion, err := semver.NewVersion("0.8.5")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(*minimumVersion) {
		return errors.New("Terraform version must be at least v0.8.5")
	}

	// This shouldn't fail, so there is no test for capturing the error.
	blacklistedVersion, err := semver.NewVersion("0.9.0")
	if err != nil {
		return err
	}

	if currentVersion.Equal(*blacklistedVersion) {
		return errors.New("Version 0.9.0 of terraform is incompatible with bbl, please try a later version.")
	}

	return nil
}

func (m Manager) Apply(bblState storage.State) (storage.State, error) {
	var err error

	if bblState.IAAS == "aws" {
		m.logger.Step("validating whether stack needs to be migrated")
		bblState, err = m.stackMigrator.Migrate(bblState)

		switch err.(type) {
		case executorError:
			return storage.State{}, NewManagerError(bblState, err.(executorError))
		case error:
			return storage.State{}, err
		}
	}

	m.logger.Step("generating terraform template")
	template := m.templateGenerator.Generate(bblState)

	m.logger.Step("generating terraform variables")
	input, err := m.inputGenerator.Generate(bblState)
	if err != nil {
		return storage.State{}, err
	}

	m.logger.Step("applying terraform template")
	tfState, err := m.executor.Apply(
		input,
		template,
		bblState.TFState,
	)

	bblState.LatestTFOutput = readAndReset(m.terraformOutputBuffer)

	switch err.(type) {
	case executorError:
		return storage.State{}, NewManagerError(bblState, err.(executorError))
	case error:
		return storage.State{}, err
	}

	bblState.TFState = tfState
	return bblState, nil
}

func (m Manager) Destroy(bblState storage.State) (storage.State, error) {
	m.logger.Step("destroying infrastructure")
	if bblState.TFState == "" {
		return bblState, nil
	}

	template := m.templateGenerator.Generate(bblState)

	input, err := m.inputGenerator.Generate(bblState)
	if err != nil {
		return storage.State{}, err
	}

	tfState, err := m.executor.Destroy(
		input,
		template,
		bblState.TFState)

	bblState.LatestTFOutput = readAndReset(m.terraformOutputBuffer)

	switch err.(type) {
	case executorError:
		return storage.State{}, NewManagerError(bblState, err.(executorError))
	case error:
		return storage.State{}, err
	}
	m.logger.Step("finished destroying infrastructure")

	bblState.TFState = tfState
	return bblState, nil
}

func (m Manager) GetOutputs(state storage.State) (map[string]interface{}, error) {
	return m.outputGenerator.Generate(state.TFState)
}

func readAndReset(buf *bytes.Buffer) string {
	contents := buf.Bytes()
	buf.Reset()

	return string(contents)
}
