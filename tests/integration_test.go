package tests

import (
	"bytes"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"github.com/Moleus/comp-arch-lab3/pkg/machine"
	translator2 "github.com/Moleus/comp-arch-lab3/pkg/translator"
	"gopkg.in/yaml.v3"
	"gotest.tools/v3/golden"
	"os"
	"strings"
	"testing"
)

type GoldenContents struct {
	TranslatorInput  string `yaml:"translator_input"`
	TranslatorOutput string `yaml:"translator_output"`
	MachineInput     string `yaml:"stdin"`
	MachineStdout    string `yaml:"stdout"`
	MachineLog       string `yaml:"log"`
}

func TestTranslationAndSimulation(t *testing.T) {
	dir, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range dir {
		t.Run(file.Name(), func(t *testing.T) {
			goldenFile := file.Name()
			runTest(t, goldenFile)
		})
	}
}

func parseGoldenFile(t *testing.T, filename string) GoldenContents {
	inputContent, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatal(err)
	}

	input := GoldenContents{}
	err = yaml.Unmarshal(inputContent, &input)
	if err != nil {
		t.Fatal(err)
	}
	return input
}

func runTest(t *testing.T, goldenFile string) {
	goldenContents := parseGoldenFile(t, goldenFile)

	translator := translator2.NewTranslator()
	program, err := translator.Translate(goldenContents.TranslatorInput)
	if err != nil {
		t.Fatal(err)
	}

	serializedMachineCode, err := isa.SerializeCode(program)
	if err != nil {
		t.Fatal(err)
	}

	ioData, err := isa.ReadIoData(strings.NewReader(goldenContents.MachineInput))
	if err != nil {
		t.Fatal(err)
	}

	dataPathOutputBuffer := bytes.NewBuffer([]byte{})
	controlUnitStateOutputBuffer := bytes.NewBuffer([]byte{})

	err = machine.RunSimulation(ioData, program, dataPathOutputBuffer, controlUnitStateOutputBuffer)
	if err != nil {
		t.Fatal(err)
	}

	testOutput := GoldenContents{
		TranslatorInput:  goldenContents.TranslatorInput,
		TranslatorOutput: string(serializedMachineCode),
		MachineInput:     goldenContents.MachineInput,
		MachineStdout:    dataPathOutputBuffer.String(),
		MachineLog:       controlUnitStateOutputBuffer.String(),
	}

	yamlOutput, err := yaml.Marshal(testOutput)
	if err != nil {
		t.Fatal(err)
	}
	golden.Assert(t, string(yamlOutput), goldenFile)
}
