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
}

type executor interface {
	Version() (string, error)
	Destroy(inputs map[string]string, terraformTemplate, tfState string) (string, error)
	Apply(inputs map[string]string, terraformTemplate, tfState string) (string, error)
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
}

func NewManager(args NewManagerArgs) Manager {
	return Manager{
		executor:              args.Executor,
		templateGenerator:     args.TemplateGenerator,
		inputGenerator:        args.InputGenerator,
		outputGenerator:       args.OutputGenerator,
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

func (m Manager) Apply(bblState storage.State) (storage.State, error) {
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
