package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	t "github.com/Moleus/comp-arch-lab3/pkg/translator"
)

var (
	inputFile  = flag.String("input", "", "Input file with assembly code (stdin if not specified)")
	targetFile = flag.String("target", "", "Target file for machine code (stdout if not specified)")
	// flags
)

func readAssemblyCode(inputFile string) ([]byte, error) {
	if inputFile == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(inputFile)
}

func writeMachineCode(machineCode []byte, targetFile string) error {
	if targetFile == "" {
		_, err := io.Copy(os.Stdout, bytes.NewReader(machineCode))
		return err
	}
	return os.WriteFile(targetFile, machineCode, 0644)
}

func main() {
	flag.Parse()

	var assemblyCode []byte

	assemblyCode, err := readAssemblyCode(*inputFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while reading input file: %s", err.Error())
		os.Exit(1)
	}

	translator := t.NewTranslator()
	translationOutput, err := translator.Translate(string(assemblyCode))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while translating assembly code: %s", err.Error())
		os.Exit(1)
	}

	serializationOutput, err := isa.SerializeCode(translationOutput)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while serializing machine code: %s", err.Error())
		os.Exit(1)
	}

	err = writeMachineCode(serializationOutput, *targetFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while writing machine code: %s", err.Error())
		os.Exit(1)
	}
}
