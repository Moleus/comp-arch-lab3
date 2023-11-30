/*
Machine
представляет модель процессора.
Включает ControlUnit и DataPath

Принимает машинный код и запускает симуляцию

ControlUnit и DataPath находятся в отдельных файлах
*/

package machine

import (
	"flag"
	"github.com/Moleus/comp-arch-lab3/cmd/translator"
	"github.com/Moleus/comp-arch-lab3/pkg/isa"
	"log"
	"os"
)

var (
	codeFilename      = flag.String("code", "", "machine code file")
	dataInputFilename = flag.String("data", "", "data input file")
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

func simulation(machineCode []translator.MachineCodeTerm, dataInput string) SimulationStatistics {
	dataPath := NewDataPath(dataInput)
	controlUnit := NewControlUnit(dataPath, machineCode)
	instructionCounter := 0

	for {
		err := controlUnit.decodeAndExecuteInstruction()
		if err != nil {
			log.Println(err)
			break
		}
		instructionCounter++
		controlUnit.PrintState()
	}

	log.Println("simulation finished")
	return SimulationStatistics{
		programOutput:      dataPath.ReadOutput(),
		instructionCounter: controlUnit.InstructionCounter,
		currentTick:        controlUnit.CurrentTick,
	}
}

func main() {
	flag.Parse()
	if *codeFilename == "" || *dataInputFilename == "" {
		flag.Usage()
		log.Fatalln("code and data filenames must be specified")
	}

	f, err := os.Open(*codeFilename)
	if err != nil {
		log.Fatalln(err)
	}

	df, err := os.Open(*dataInputFilename)
	if err != nil {
		log.Fatalln(err)
	}

	machineCode, err := isa.ReadCode(f)
	if err != nil {
		log.Fatalln(err)
	}

	var dataInput []byte
	_, err = df.Read(dataInput)
	if err != nil {
		log.Fatalln(err)
	}

	simulation(machineCode, string(dataInput))
}
