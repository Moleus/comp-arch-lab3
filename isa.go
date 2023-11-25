/*
ISA - Instruction Set Architecture (Система команд)
Фон-Неймановская архитектура

Задачи:
- читает машинный код из файла
- записывает машинный код в файл

По сути занимается сериализацией и десериализаций программы в JSON

*/
package isa

import (
  "fmt"
  "io"
  "log"
)

/* accumulator based ISA */

type AddrInstr int
type NoAddrInstr int

const (
  AddrInstrHalt AddrInstr = iota


