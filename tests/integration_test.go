package tests

import (
	"bytes"
	"github.com/Moleus/comp-arch-lab3/cmd/simulation"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"github.com/Moleus/comp-arch-lab3/pkg/machine"
	translator2 "github.com/Moleus/comp-arch-lab3/pkg/translator"
	"github.com/gkampitakis/go-snaps/snaps"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"strings"
	"testing"
)

type TestInput struct {
	TranslatorInput string `yaml:"translator_input"`
	MachineInput    string `yaml:"machine_input"`
}

type TestOutput struct {
	translatorOutput string `yaml:"translator_output"`
	machineStdout    string `yaml:"stdout"`
	machineLog       string `yaml:"log"`
}

func TestTranslationAndSimulation(t *testing.T) {
	dir, err := os.ReadDir("inputs")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range dir {
		t.Run(file.Name(), func(t *testing.T) {
			input := parseInputFile(t, file.Name())
			runTest(t, input)
		})
	}
}

func parseInputFile(t *testing.T, filename string) TestInput {
	inputContent, err := os.ReadFile("inputs/" + filename)
	if err != nil {
		t.Fatal(err)
	}

	input := TestInput{}
	err = yaml.Unmarshal(inputContent, &input)
	if err != nil {
		t.Fatal(err)
	}
	return input
}

func runTest(t *testing.T, input TestInput) {
	// translate input to machineCode
	translator := translator2.NewTranslator()
	machineCode, err := translator.Translate(input.TranslatorInput)
	if err != nil {
		t.Fatal(err)
	}
	serializedMachineCode, err := isa.SerializeCode(machineCode)
	if err != nil {
		t.Fatal(err)
	}

	ioData, err := isa.ReadIoData(strings.NewReader(input.MachineInput))
	if err != nil {
		t.Fatal(err)
	}

	logOutputBuffer := bytes.NewBuffer([]byte{})
	dataPathOutputBuffer := bytes.NewBuffer([]byte{})

	clock := simulation.Clock{CurrentTick: 0}
	logger := simulation.InitLogger(logOutputBuffer, &clock, slog.LevelDebug)

	machine.RunSimulation(ioData, machineCode, logger, dataPathOutputBuffer)

	testOutput := TestOutput{
		translatorOutput: string(serializedMachineCode),
		machineStdout:    input.MachineInput,
		machineLog:       logOutputBuffer.String(),
	}

	snaps.MatchSnapshot(t, testOutput)
}
