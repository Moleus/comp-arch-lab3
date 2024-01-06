package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"github.com/Moleus/comp-arch-lab3/pkg/machine"
	log "github.com/Moleus/comp-arch-lab3/pkg/logging"
)

var (
  logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
  programCodeFilename = flag.String("program", "", "Path to program file")
  dataInputFilename = flag.String("io-data", "", "Path to IO data file")
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
    fmt.Sprintln(os.Stderr, "Program file is not specified")
    flag.Usage()
  }

  if *dataInputFilename == "" {
    fmt.Sprintln(os.Stderr, "IO data file is not specified")
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

	machineCode, err := isa.ReadCode(f)
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
	clock := &Clock{CurrentTick: 0}
  logLevel := log.ParseLogLevel(*logLevel)
	defaultHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(log.NewTickLoggerHandler(defaultHandler, clock))

  machine.RunSimulation(ioData, machineCode, logger)
  // process simulation results
}
