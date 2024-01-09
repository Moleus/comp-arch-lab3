/*
Machine
представляет модель процессора.
Включает ControlUnit и DataPath

Принимает машинный код и запускает симуляцию

ControlUnit и DataPath находятся в отдельных файлах
*/

package machine

import (
	"fmt"
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

func RunSimulation(dataInput []isa.IoData, program isa.Program, dataPathOutput io.Writer, controlUnitStateOutput io.Writer) {
	clock := &Clock{currentTick: 0}
	datapath := NewDataPath(dataInput, dataPathOutput, clock)
	controlUnit := NewControlUnit(program, datapath, controlUnitStateOutput, clock)

	log.Println("starting simulation")

	err := controlUnit.RunInstructionCycle()
	if err != nil {
		// TODO: ignore NOP errors
		fmt.Print(err)
	}

	log.Println("simulation finished")
}
