/*
Machine
представляет модель процессора.
Включает ControlUnit и DataPath

Принимает машинный код и запускает симуляцию

ControlUnit и DataPath находятся в отдельных файлах
*/

package machine

import (
	"log"
	"log/slog"

	"github.com/Moleus/comp-arch-lab3/pkg/isa"
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

func RunSimulation(dataInput []isa.IoData, program []isa.MachineCodeTerm, logger *slog.Logger) SimulationStatistics {
	datapath := NewDataPath(dataInput)
	controlUnit := NewControlUnit(program, datapath, logger)

	log.Println("starting simulation")

	err := controlUnit.RunInstructionCycle()
	if err != nil {
		// TODO: ignore NOP errors
		logger.Error(err.Error())
	}

	log.Println("simulation finished")
	return SimulationStatistics{
		programOutput:      datapath.ReadOutput(),
		instructionCounter: controlUnit.instructionCounter,
		currentTick:        controlUnit.currentTick,
	}
}
