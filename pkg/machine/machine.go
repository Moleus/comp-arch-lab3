package machine

import (
	"errors"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"io"
	"log"
)

type Machine struct {
}

type SimulationStatistics struct {
}

func RunSimulation(dataInput []isa.IoData, program isa.Program, dataPathOutput io.Writer, controlUnitStateOutput io.Writer) error {
	clock := &Clock{currentTick: 0}
	dataPath := NewDataPath(dataInput, dataPathOutput, clock)
	controlUnit := NewControlUnit(program, dataPath, controlUnitStateOutput, clock)

	log.Println("starting simulation")

	controlUnit.PresetInstructionCounter(controlUnit.program.StartAddress)
	err := controlUnit.RunInstructionCycle()
	var controlUnitError *ControlUnitError
	if err == nil {
		return errors.New("simulation should finish with HLT")
	} else if !errors.As(err, &controlUnitError) {
		return err
	}

	log.Printf("simulation finished. Instructions executed: %d, ticks: %d", controlUnit.ExecutedInstructions, clock.GetCurrentTick())
	return nil
}
