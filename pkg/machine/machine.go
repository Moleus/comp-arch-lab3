/*
Machine
представляет модель процессора.
Включает ControlUnit и DataPath

Принимает машинный код и запускает симуляцию

ControlUnit и DataPath находятся в отдельных файлах
*/

package machine

import (
	"errors"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"io"
	"log"
)

type Machine struct {
	dataPath    DataPath
	controlUnit ControlUnit
}

type SimulationStatistics struct {
	programOutput      string
	instructionCounter int
	currentTick        int
}

func RunSimulation(dataInput []isa.IoData, program isa.Program, dataPathOutput io.Writer, controlUnitStateOutput io.Writer) error {
	clock := &Clock{currentTick: 0}
	datapath := NewDataPath(dataInput, dataPathOutput, clock)
	controlUnit := NewControlUnit(program, datapath, controlUnitStateOutput, clock)

	log.Println("starting simulation")

	controlUnit.PresetInstructionCounter(controlUnit.program.StartAddress)
	err := controlUnit.RunInstructionCycle()
	var controlUnitError *ControlUnitError
	if err == nil {
		return errors.New("simulation should finish with HLT")
	} else if !errors.As(err, &controlUnitError) {
		return err
	}

	log.Println("simulation finished")
	return nil
}
