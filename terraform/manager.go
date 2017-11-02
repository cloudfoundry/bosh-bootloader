package terraform

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/coreos/go-semver/semver"
)

type Manager struct {
	executor              executor
	templateGenerator     TemplateGenerator
	inputGenerator        InputGenerator
	terraformOutputBuffer *bytes.Buffer
	logger                logger
}

type executor interface {
	Version() (string, error)
	Destroy(inputs map[string]interface{}) error
	IsInitialized() bool
	Init(terraformTemplate string, inputs map[string]interface{}) error
	Apply() error
	Outputs() (map[string]interface{}, error)
	Output(string) (string, error)
}

type InputGenerator interface {
	Generate(storage.State) (map[string]interface{}, error)
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
	TerraformOutputBuffer *bytes.Buffer
	Logger                logger
}

func NewManager(args NewManagerArgs) Manager {
	return Manager{
		executor:              args.Executor,
		templateGenerator:     args.TemplateGenerator,
		inputGenerator:        args.InputGenerator,
		terraformOutputBuffer: args.TerraformOutputBuffer,
		logger:                args.Logger,
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

	minimumVersion, err := semver.NewVersion("0.10.0")
	if err != nil {
		return err
	}

	if currentVersion.LessThan(*minimumVersion) {
		return errors.New("Terraform version must be at least v0.10.0")
	}

	return nil
}

func (m Manager) IsInitialized() bool {
	return m.executor.IsInitialized()
}

func (m Manager) Init(bblState storage.State) error {
	m.logger.Step("generating terraform template")
	template := m.templateGenerator.Generate(bblState)

	m.logger.Step("generating terraform variables")
	input, err := m.inputGenerator.Generate(bblState)
	if err != nil {
		return fmt.Errorf("Input generator generate: %s", err)
	}

	err = m.executor.Init(template, input)
	if err != nil {
		return fmt.Errorf("Executor init: %s", err)
	}

	return nil
}

func (m Manager) Apply(bblState storage.State) (storage.State, error) {
	m.logger.Step("terraform apply")
	err := m.executor.Apply()

	bblState.LatestTFOutput = readAndReset(m.terraformOutputBuffer)

	switch err.(type) {
	case executorError:
		return storage.State{}, NewManagerError(bblState, err.(executorError))
	case error:
		return storage.State{}, err
	}

	return bblState, nil
}

func (m Manager) Destroy(bblState storage.State) (storage.State, error) {
	m.logger.Step("destroying infrastructure")
	m.logger.Step("generating terraform variables")
	input, err := m.inputGenerator.Generate(bblState)
	if err != nil {
		return storage.State{}, fmt.Errorf("Input generator generate: %s", err)
	}

	m.logger.Step("terraform destroy")
	err = m.executor.Destroy(input)
	bblState.LatestTFOutput = readAndReset(m.terraformOutputBuffer)

	switch err.(type) {
	case executorError:
		return storage.State{}, NewManagerError(bblState, err.(executorError))
	case error:
		return storage.State{}, fmt.Errorf("Executor destroy: %s", err)
	}
	m.logger.Step("finished destroying infrastructure")

	return bblState, nil
}

func (m Manager) GetOutputs() (Outputs, error) {
	tfOutputs, err := m.executor.Outputs()
	if err != nil {
		return Outputs{}, err
	}

	return Outputs{Map: tfOutputs}, nil
}

func readAndReset(buf *bytes.Buffer) string {
	contents := buf.Bytes()
	buf.Reset()

	return string(contents)
}
