package tests

import (
	"bytes"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"github.com/Moleus/comp-arch-lab3/pkg/machine"
	translator2 "github.com/Moleus/comp-arch-lab3/pkg/translator"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"testing"
)

type TestInput struct {
	TranslatorInput string `yaml:"translator_input"`
	MachineInput    string `yaml:"machine_input"`
}

type TestOutput struct {
	TranslatorOutput string `yaml:"translator_output"`
	MachineStdout    string `yaml:"stdout"`
	MachineLog       string `yaml:"log"`
}

func TestTranslationAndSimulation(t *testing.T) {
	dir, err := os.ReadDir("inputs")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range dir {
		t.Run(file.Name(), func(t *testing.T) {
			input := parseInputFile(t, file.Name())
			goldenFile := "golden/" + file.Name()
			runTest(t, input, goldenFile)
		})
	}
}

func TestSimplePlusProgram(t *testing.T) {
	input := parseInputFile(t, "plus.yml")
	goldenFilename := "golden/plus.yml"
	runTest(t, input, goldenFilename)
}

func TestSimpleStack(t *testing.T) {
	input := parseInputFile(t, "stack.yml")
	goldenFilename := "golden/stack.yml"
	runTest(t, input, goldenFilename)
}

func TestCatIOInterrupts(t *testing.T) {
	input := parseInputFile(t, "cat.yml")
	goldenFilename := "golden/cat.yml"
	runTest(t, input, goldenFilename)
}

func TestHelloOutput(t *testing.T) {
	input := parseInputFile(t, "hello.yml")
	goldenFilename := "golden/hello.yml"
	runTest(t, input, goldenFilename)
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

func runTest(t *testing.T, input TestInput, goldenFile string) {
	// translate input to program
	translator := translator2.NewTranslator()
	program, err := translator.Translate(input.TranslatorInput)
	if err != nil {
		t.Fatal(err)
	}
	serializedMachineCode, err := isa.SerializeCode(program)
	if err != nil {
		t.Fatal(err)
	}

	ioData, err := isa.ReadIoData(strings.NewReader(input.MachineInput))
	if err != nil {
		t.Fatal(err)
	}

	dataPathOutputBuffer := bytes.NewBuffer([]byte{})
	controlUnitStateOutputBuffer := bytes.NewBuffer([]byte{})

	err = machine.RunSimulation(ioData, program, dataPathOutputBuffer, controlUnitStateOutputBuffer)
	if err != nil {
		t.Fatal(err)
	}

	testOutput := TestOutput{
		TranslatorOutput: string(serializedMachineCode),
		MachineStdout:    dataPathOutputBuffer.String(),
		MachineLog:       controlUnitStateOutputBuffer.String(),
	}

	yamlOutput, err := yaml.Marshal(testOutput)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(goldenFile, yamlOutput, 0644); err != nil {
		t.Fatal(err)
	}
}
