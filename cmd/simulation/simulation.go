package simulation

import (
	"flag"
	"fmt"
	"os"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"github.com/Moleus/comp-arch-lab3/pkg/machine"
)

var (
	logLevel            = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	programCodeFilename = flag.String("program", "", "Path to program file")
	dataInputFilename   = flag.String("io-data", "", "Path to IO data file")
)

type Clock struct {
	CurrentTick int
}

func (c *Clock) GetCurrentTick() int {
	return c.CurrentTick
}

func main() {
	flag.Parse()

	if *programCodeFilename == "" {
		fmt.Fprintln(os.Stderr, "Program file is not specified")
		flag.Usage()
	}

	if *dataInputFilename == "" {
		fmt.Fprintln(os.Stderr, "IO data file is not specified")
		flag.Usage()
	}

	f, err := os.Open(*programCodeFilename)
	// print error and exit
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening program file: %s", err.Error())
		os.Exit(1)
	}

	df, err := os.Open(*dataInputFilename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while opening IO data file: %s", err.Error())
		os.Exit(1)
	}

	program, err := isa.ReadCode(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading program file: %s", err.Error())
		os.Exit(1)
	}

	ioData, err := isa.ReadIoData(df)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading IO data file: %s", err.Error())
		os.Exit(1)
	}

	// TODO: read program, read data, flags etc
	dataPathOutput := os.Stdout
	controlUnitStateOutput := os.Stdout

	machine.RunSimulation(ioData, program, dataPathOutput, controlUnitStateOutput)
	// process simulation results
}
