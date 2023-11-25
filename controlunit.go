/*
ControlUnit.
Интерпретирует команды.
Управляющие потоки идут в ControlUnit
hw - hardwired. Реализуется как часть модели. microcode не нужен.

На вход получает информацию, на выходе выставляет сигналы. Возможно state register и не нужен...

У ControlUnit должно быть состояние, которое описывает текущее состояние исполнения команды (методичка)

Потактовое исполнение команд.
Цикл команды (стр. 53):
1. Цикл выборки команды (Instruction Fetch)
2. Цикл выборки адреса (Address Fetch)
3. Цикл выборки операнда (Operand Fetch)
4. Цикл исполнения (Execution)
5. Цикл прерывания (Interruption) - нужен для ввода-вывода
*/