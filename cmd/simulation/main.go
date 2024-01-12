package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"github.com/Moleus/comp-arch-lab3/pkg/machine"
)

var (
	programCodeFilename = flag.String("program", "", "Path to program file")
	dataInputFilename   = flag.String("io-data", "", "Path to IO data file")
)

func main() {
	flag.Parse()

	if *programCodeFilename == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Program file is not specified")
		flag.Usage()
	}

	if *dataInputFilename == "" {
		_, _ = fmt.Fprintln(os.Stderr, "IO data file is not specified")
		flag.Usage()
	}

	f, err := os.Open(*programCodeFilename)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while opening program file: %s", err.Error())
		os.Exit(1)
	}

	df, err := os.Open(*dataInputFilename)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while opening IO data file: %s", err.Error())
		os.Exit(1)
	}

	program, err := isa.ReadCode(f)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while reading program file: %s", err.Error())
		os.Exit(1)
	}

	ioData, err := isa.ReadIoData(df)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while reading IO data file: %s", err.Error())
		os.Exit(1)
	}

	dataPathOutput := os.Stdout
	controlUnitStateOutput := os.Stdout

	err = machine.RunSimulation(ioData, program, dataPathOutput, controlUnitStateOutput)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error while running simulation: %s", err.Error())
		os.Exit(1)
	}
}
