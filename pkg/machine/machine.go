/*
Machine
представляет модель процессора.
Включает ControlUnit и DataPath

Принимает машинный код и запускает симуляцию

ControlUnit и DataPath находятся в отдельных файлах
*/

package machine

import (
	"io"
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

func RunSimulation(dataInput []isa.IoData, program []isa.MachineCodeTerm, logger *slog.Logger, dataPathOutput io.Writer) {
	datapath := NewDataPath(dataInput, dataPathOutput)
	controlUnit := NewControlUnit(program, datapath, logger)

	log.Println("starting simulation")

	err := controlUnit.RunInstructionCycle()
	if err != nil {
		// TODO: ignore NOP errors
		logger.Error(err.Error())
	}

	log.Println("simulation finished")
}
