package main

import (
	"bytes"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	log "github.com/Moleus/comp-arch-lab3/pkg/logging"
	"github.com/Moleus/comp-arch-lab3/pkg/machine"
)

var (
  logLevel = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
)

func simulation(logger *slog.Logger) {
  // TODO: implement
  dataInput := bytes.NewBuffer([]byte{})
  program := []isa.MachineCodeTerm{}

  datapath := machine.NewDataPath(*dataInput)
  controlUnit := machine.NewControlUnit(program, datapath, logger)
  err := controlUnit.RunInstructionCycle()
  if err != nil {
    // TODO: ignore NOP errors
    logger.Error(err.Error())
  }
}

func parseLogLevel(level string) slog.Level {
  switch level {
  case "debug":
    return slog.LevelDebug
  case "info":
    return slog.LevelInfo
  case "warn":
    return slog.LevelWarn
  case "error":
    return slog.LevelError
  default:
    panic(fmt.Sprintf("Unknown log level %s", level))
  }
}

type Clock struct {
  CurrentTick int
}

func (c *Clock) GetCurrentTick() int {
  return c.CurrentTick
}

func main() {
  // TODO: read program, read data, flags etc
	clock := &Clock{CurrentTick: 0}
  logLevel := parseLogLevel(*logLevel)
	defaultHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	logger := slog.New(log.NewTickLoggerHandler(defaultHandler, clock))

  simulation(logger)
  // process simulation results
}
