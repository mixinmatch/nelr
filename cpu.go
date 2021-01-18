package main
import (

)

type Cpu struct {
	memory Memory
	PC     uint16
	A      byte
	X      byte
	Y      byte
	SP     byte

	P byte
	cycles uint64
	suspendCycles uint64

	nmiRequested bool
	irqRequested bool
}

const (
	_   = iota
	imp = "implicit"
	acc = "accumulator"
	imm = "immediate"
	zep = "zeroPage"
	zpx = "zeroPageX"
	zpy = "zeroPageY"
	rel = "relative"
	abs = "absolute"
	abx = "absoluteX"
	aby = "absoluteY"
	ind = "indirect"
	inx = "indexedIndirect"
	iny = "indirectIndexed"
)

const (
	CFlag = 1<<0 //0x01
	ZFlag = 1<<1 //0x02
	IFlag = 1<<2 //0x04
	DFlag = 1<<3 //0x08
	BFlag = 1<<4 //0x10
	VFlag = 1<<6 //0x40
	NFlag = 1<<7 //0x80
)

var opcodes = [256]struct {
	name             string
	addressingMode   string
	cycles           uint16
	additionalCycles uint16
	size             byte
}{
	{"BRK", imp, 7, 0, 1}, //0x0
	{"ORA", inx, 6, 0, 2}, // x1
	{},                    //{STP, imp, 0, 0, 0}, // x2
	{},                    //{SLO, inx, 8, 0, 0}, // x3
	{"NOP", zep, 3, 0, 2}, // x4
	{"ORA", zep, 3, 0, 2}, // x5
	{"ASL", zep, 5, 0, 2}, // x6
	{},                    //{SLO, zep, 5, 0, 0}, // x7
	{"PHP", imp, 3, 0, 1}, // x8
	{"ORA", imm, 2, 0, 2}, // x9
	{"ASL", acc, 2, 0, 1}, // xA
	{},                    //{ANC, imm, 2, 0}, // xB
	{"NOP", abs, 4, 0, 3}, // xC
	{"ORA", abs, 4, 0, 3}, // xD
	{"ASL", abs, 6, 0, 3}, // xE
	{},                    //{SLO, abs, 6, 0}, // xF

	// 1x
	{"BPL", rel, 2, 1, 2}, // x0
	{"ORA", iny, 5, 1, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{SLO, iny, 8, 0}, // x3
	{"NOP", zpx, 4, 0, 2}, // x4
	{"ORA", zpx, 4, 0, 2}, // x5
	{"ASL", zpx, 6, 0, 2}, // x6
	{},                    //{SLO, zpx, 6, 0}, // x7
	{"CLC", imp, 2, 0, 1}, // x8
	{"ORA", aby, 4, 1, 3}, // x9
	{"NOP", imp, 2, 0, 1}, // xA
	{},                    //{SLO, aby, 7, 0}, // xB
	{"NOP", abx, 4, 1, 3}, // xC
	{"ORA", abx, 4, 1, 3}, // xD
	{"ASL", abx, 7, 0, 3}, // xE
	{},                    //{SLO, abx, 7, 0}, // xF

	// 2x
	{"JSR", abs, 6, 0, 3}, // x0
	{"AND", inx, 6, 0, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{RLA, inx, 8, 0}, // x3
	{"BIT", zep, 3, 0, 2}, // x4
	{"AND", zep, 3, 0, 2}, // x5
	{"ROL", zep, 5, 0, 2}, // x6
	{},                    //{RLA, zep, 5, 0}, // x7
	{"PLP", imp, 4, 0, 1}, // x8
	{"AND", imm, 2, 0, 2}, // x9
	{"ROL", acc, 2, 0, 1}, // xA
	{},                    //{ANC, imm, 2, 0}, // xB
	{"BIT", abs, 4, 0, 3}, // xC
	{"AND", abs, 4, 0, 3}, // xD
	{"ROL", abs, 6, 0, 3}, // xE
	{},                    //{RLA, abs, 6, 0}, // xF

	// 3x
	{"BMI", rel, 2, 1, 2}, // x0
	{"AND", iny, 5, 1, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{RLA, iny, 8, 0}, // x3
	{"NOP", zpx, 4, 0, 2}, // x4
	{"AND", zpx, 4, 0, 2}, // x5
	{"ROL", zpx, 6, 0, 2}, // x6
	{},                    //{RLA, zpx, 6, 0}, // x7
	{"SEC", imp, 2, 0, 1}, // x8
	{"AND", aby, 4, 1, 3}, // x9
	{"NOP", imp, 2, 0, 1}, // xA
	{},                    //{RLA, aby, 7, 0}, // xB
	{"NOP", abx, 4, 1, 3}, // xC
	{"AND", abx, 4, 1, 3}, // xD
	{"ROL", abx, 7, 0, 3}, // xE
	{},                    //{RLA, abx, 7, 0}, // xF

	// 4x
	{"RTI", imp, 6, 0, 1}, // x0
	{"EOR", inx, 6, 0, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{SRE, inx, 8, 0}, // x3
	{"NOP", zep, 3, 0, 2}, // x4
	{"EOR", zep, 3, 0, 2}, // x5
	{"LSR", zep, 5, 0, 2}, // x6
	{},                    //{SRE, zep, 5, 0}, // x7
	{"PHA", imp, 3, 0, 1}, // x8
	{"EOR", imm, 2, 0, 2}, // x9
	{"LSR", imp, 2, 0, 1}, // xA
	{},                    //{ALR, imm, 2, 0}, // xB
	{"JMP", abs, 3, 0, 3}, // xC
	{"EOR", abs, 4, 0, 3}, // xD
	{"LSR", abs, 6, 0, 3}, // xE
	{},                    //{SRE, abs, 6, 0}, // xF

	// 5x
	{"BVC", rel, 2, 1, 2}, // x0
	{"EOR", iny, 5, 1, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{SRE, iny, 8, 0}, // x3
	{"NOP", zpx, 4, 0, 2}, // x4
	{"EOR", zpx, 4, 0, 2}, // x5
	{"LSR", zpx, 6, 0, 2}, // x6
	{},                    //{SRE, zpx, 6, 0}, // x7
	{"CLI", imp, 2, 0, 1}, // x8
	{"EOR", aby, 4, 1, 3}, // x9
	{"NOP", imp, 2, 0, 1}, // xA
	{},                    //{SRE, aby, 7, 0}, // xB
	{"NOP", abx, 4, 1, 3}, // xC
	{"EOR", abx, 4, 1, 3}, // xD
	{"LSR", abx, 7, 0, 3}, // xE
	{},                    //{SRE, abx, 7, 0}, // xF

	// 6x
	{"RTS", imp, 6, 0, 1}, // x0
	{"ADC", inx, 6, 0, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{RRA, inx, 8, 0}, // x3
	{"NOP", zep, 3, 0, 2}, // x4
	{"ADC", zep, 3, 0, 2}, // x5
	{"ROR", zep, 5, 0, 2}, // x6
	{},                    //{RRA, zep, 5, 0}, // x7
	{"PLA", imp, 4, 0, 1}, // x8
	{"ADC", imm, 2, 0, 2}, // x9
	{"ROR", imp, 2, 0, 1}, // xA
	{},                    //{ARR, imm, 2, 0}, // xB
	{"JMP", ind, 5, 0, 3}, // xC
	{"ADC", abs, 4, 0, 3}, // xD
	{"ROR", abs, 6, 0, 3}, // xE
	{},                    //{RRA, abs, 6, 0}, // xF

	// 7x
	{"BVS", rel, 2, 1, 2}, // x0
	{"ADC", iny, 5, 1, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{RRA, iny, 8, 0}, // x3
	{"NOP", zpx, 4, 0, 2}, // x4
	{"ADC", zpx, 4, 0, 2}, // x5
	{"ROR", zpx, 6, 0, 2}, // x6
	{},                    //{RRA, zpx, 6, 0}, // x7
	{"SEI", imp, 2, 0, 1}, // x8
	{"ADC", aby, 4, 1, 3}, // x9
	{"NOP", imp, 2, 0, 1}, // xA
	{},                    //{RRA, aby, 7, 0}, // xB
	{"NOP", abx, 4, 1, 3}, // xC
	{"ADC", abx, 4, 1, 3}, // xD
	{"ROR", abx, 7, 0, 3}, // xE
	{},                    //{RRA, abx, 7, 0}, // xF

	// 8x
	{"NOP", imm, 2, 0, 2}, // x0
	{"STA", inx, 6, 0, 2}, // x1
	{"NOP", imm, 2, 0, 2}, // x2
	{},                    //{SAX, inx, 6, 0}, // x3
	{"STY", zep, 3, 0, 2}, // x4
	{"STA", zep, 3, 0, 2}, // x5
	{"STX", zep, 3, 0, 2}, // x6
	{},                    //{SAX, zep, 3, 0}, // x7
	{"DEY", imp, 2, 0, 1}, // x8
	{"NOP", imm, 2, 0, 2}, // x9
	{"TXA", imp, 2, 0, 1}, // xA
	{},                    //{XAA, imm, 2, 1}, // xB
	{"STY", abs, 4, 0, 3}, // xC
	{"STA", abs, 4, 0, 3}, // xD
	{"STX", abs, 4, 0, 3}, // xE
	{},                    //{SAX, abs, 4, 0}, // xF

	// 9x
	{"BCC", rel, 2, 1, 2}, // x0
	{"STA", iny, 6, 0, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{AHX, iny, 6, 0}, // x3
	{"STY", zpx, 4, 0, 2}, // x4
	{"STA", zpx, 4, 0, 2}, // x5
	{"STX", zpy, 4, 0, 2}, // x6
	{},                    //{SAX, zpy, 4, 0}, // x7
	{"TYA", imp, 2, 0, 1}, // x8
	{"STA", aby, 5, 0, 3}, // x9
	{"TXS", imp, 2, 0, 1}, // xA
	{},                    //{TAS, aby, 5, 0}, // xB
	{},                    //{SHY, abx, 5, 0}, // xC
	{"STA", abx, 5, 0, 3}, // xD
	{},                    //{SHX, aby, 5, 0}, // xE
	{},                    //{AHX, aby, 5, 0}, // xF

	// Ax
	{"LDY", imm, 2, 0, 2}, // x0
	{"LDA", inx, 6, 0, 2}, // x1
	{"LDX", imm, 2, 0, 2}, // x2
	{},                    //{LAX, inx, 6, 0}, // x3
	{"LDY", zep, 3, 0, 2}, // x4
	{"LDA", zep, 3, 0, 2}, // x5
	{"LDX", zep, 3, 0, 2}, // x6
	{},                    //{LAX, zep, 3, 0}, // x7
	{"TAY", imp, 2, 0, 1}, // x8
	{"LDA", imm, 2, 0, 2}, // x9
	{"TAX", imp, 2, 0, 1}, // xA
	{},                    //{LAX, imm, 2, 0}, // xB
	{"LDY", abs, 4, 0, 3}, // xC
	{"LDA", abs, 4, 0, 3}, // xD
	{"LDX", abs, 4, 0, 3}, // xE
	{},                    //{LAX, abs, 4, 0}, // xF

	// Bx
	{"BCS", rel, 2, 1, 2}, // x0
	{"LDA", iny, 5, 1, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{LAX, iny, 5, 1}, // x3
	{"LDY", zpx, 4, 0, 2}, // x4
	{"LDA", zpx, 4, 0, 2}, // x5
	{"LDX", zpy, 4, 0, 2}, // x6
	{},                    //{LAX, zpy, 4, 0}, // x7
	{"CLV", imp, 2, 0, 1}, // x8
	{"LDA", aby, 4, 1, 3}, // x9
	{"TSX", imp, 2, 0, 1}, // xA
	{},                    //{LAS, aby, 4, 1}, // xB
	{"LDY", abx, 4, 1, 3}, // xC
	{"LDA", abx, 4, 1, 3}, // xD
	{"LDX", aby, 4, 1, 3}, // xE
	{},                    //{LAX, aby, 4, 1}, // xF

	// Cx
	{"CPY", imm, 2, 0, 2}, // x0
	{"CMP", inx, 6, 0, 2}, // x1
	{"NOP", imm, 2, 0, 2}, // x2
	{},                    //{DCP, inx, 8, 0}, // x3
	{"CPY", zep, 3, 0, 2}, // x4
	{"CMP", zep, 3, 0, 2}, // x5
	{"DEC", zep, 5, 0, 2}, // x6
	{},                    //{DCP, zep, 5, 0}, // x7
	{"INY", imp, 2, 0, 1}, // x8
	{"CMP", imm, 2, 0, 2}, // x9
	{"DEX", imp, 2, 0, 1}, // xA
	{},                    //{AXS, imm, 2, 0}, // xB
	{"CPY", abs, 4, 0, 3}, // xC
	{"CMP", abs, 4, 0, 3}, // xD
	{"DEC", abs, 6, 0, 3}, // xE
	{},                    //{DCP, abs, 6, 0}, // xF

	// Dx
	{"BNE", rel, 2, 1, 2}, // x0
	{"CMP", iny, 5, 1, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{DCP, iny, 8, 0}, // x3
	{"NOP", zpx, 4, 0, 2}, // x4
	{"CMP", zpx, 4, 0, 2}, // x5
	{"DEC", zpx, 6, 0, 2}, // x6
	{},                    //{DCP, zpx, 6, 0}, // x7
	{"CLD", imp, 2, 0, 1}, // x8
	{"CMP", aby, 4, 1, 3}, // x9
	{"NOP", imp, 2, 0, 1}, // xA
	{},                    //{DCP, aby, 7, 0}, // xB
	{"NOP", abx, 4, 1, 3}, // xC
	{"CMP", abx, 4, 1, 3}, // xD
	{"DEC", abx, 7, 0, 3}, // xE
	{},                    //{DCP, abx, 7, 0}, // xF

	// Ex
	{"CPX", imm, 2, 0, 2}, // x0
	{"SBC", inx, 6, 0, 2}, // x1
	{"NOP", imm, 2, 0, 2}, // x2
	{},                    //{ISC, inx, 8, 0}, // x3
	{"CPX", zep, 3, 0, 2}, // x4
	{"SBC", zep, 3, 0, 2}, // x5
	{"INC", zep, 5, 0, 2}, // x6
	{},                    //{ISC, zep, 5, 0}, // x7
	{"INX", imp, 2, 0, 1}, // x8
	{"SBC", imm, 2, 0, 2}, // x9
	{"NOP", imp, 2, 0, 1}, // xA
	{"SBC", imm, 2, 0, 0}, // xB
	{"CPX", abs, 4, 0, 3}, // xC
	{"SBC", abs, 4, 0, 3}, // xD
	{"INC", abs, 6, 0, 3}, // xE
	{},                    //{ISC, abs, 6, 0}, // xF

	// Fx
	{"BEQ", rel, 2, 1, 2}, // x0
	{"SBC", iny, 5, 1, 2}, // x1
	{},                    //{STP, imp, 0, 0}, // x2
	{},                    //{ISC, iny, 8, 0}, // x3
	{"NOP", zpx, 4, 0, 2}, // x4
	{"SBC", zpx, 4, 0, 2}, // x5
	{"INC", zpx, 6, 0, 2}, // x6
	{},                    //{ISC, zpx, 6, 0}, // x7
	{"SED", imp, 2, 0, 1}, // x8
	{"SBC", aby, 4, 1, 3}, // x9
	{"NOP", imp, 2, 0, 1}, // xA
	{},                    //{ISC, aby, 7, 0}, // xB
	{"NOP", abx, 4, 1, 3}, // xC
	{"SBC", abx, 4, 1, 3}, // xD
	{"INC", abx, 7, 0, 3}, // xE
	{},                    //{ISC, abx, 7, 0}} // xF
}

//Addressing modes

func (cpu *Cpu) immediateAddress() uint16 {
	return cpu.PC + 1
}

func (cpu *Cpu) zeroPageAddress() uint16 {
	address := cpu.memory.Read(cpu.PC + 1)
	return uint16(address)
}

func (cpu *Cpu) absoluteAddress() uint16 {
	address := cpu.ReadUint16(cpu.PC + 1)
	return address
}

func (cpu *Cpu) relativeAddress() uint16 {
	offset := uint16(cpu.memory.Read(cpu.PC + 1))

	var address uint16
	//offset is byte from 0x0 to 0xff
	//positive offset if 0 <= offset <= 127
	if offset < 128 {
		address = cpu.PC + offset + 2
	} else { //negative if 128 <= offset <= 255
		// convert uint = int + (UINT_MAX + 1)
		// address = cpu.PC + (offset - (0xFF + 1))
		address = cpu.PC + offset - (0x100) + 2
	}

	return address
}

func (cpu *Cpu) indirectAddress() uint16 {
	addr := cpu.ReadUint16(cpu.PC + 1)
	return cpu.ReadBuggyUint16(addr)
}

func (cpu *Cpu) ReadBuggyUint16(addr uint16) uint16 {

	//6502 bug causing incrementing lower byte
	//not carrying to higher byte
	lowAddress := addr
	highAddress := (addr & 0xFF00) | uint16(byte(addr)+1)

	indirectLowAddress := cpu.memory.Read(lowAddress)
	indirectHighAddress := cpu.memory.Read(highAddress)

	return (uint16(indirectHighAddress) << 8) + uint16(indirectLowAddress)
}

func (cpu *Cpu) zeroPageXAddress() uint16 {
	address := cpu.memory.Read(cpu.PC + 1)
	return uint16(address + cpu.X)
}

func (cpu *Cpu) zeroPageYAddress() uint16 {
	address := cpu.memory.Read(cpu.PC + 1)
	return uint16(address + cpu.Y)
}

func (cpu *Cpu) absoluteXAddress() uint16 {
	address := cpu.ReadUint16(cpu.PC+1) + uint16(cpu.X)
	return address
}

func (cpu *Cpu) absoluteYAddress() uint16 {
	address := cpu.ReadUint16(cpu.PC+1) + uint16(cpu.Y)
	return address
}

func (cpu *Cpu) indexIndirectAddress() uint16 {
	tmpAddress := cpu.memory.Read(cpu.PC+1) + cpu.X
	address := cpu.ReadBuggyUint16(uint16(tmpAddress))
	return address
}

func (cpu *Cpu) indirectIndexedAddress() uint16 {
	tmpAddress := cpu.memory.Read(cpu.PC + 1)
	address := cpu.ReadBuggyUint16(uint16(tmpAddress)) + uint16(cpu.Y)
	return address
}

//instructions
func (cpu *Cpu) adc(address uint16) {
	a := cpu.A
	b := cpu.memory.Read(address)
	c := cpu.getSetFlag(CFlag)

	cpu.A = a + b + c
	sum := int(a) + int(b) + int(c)

	if sum > 0xFF {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if (a^b)&NFlag == 0 && (a^cpu.A)&NFlag != 0 {
		cpu.setFlag(VFlag)
	} else {
		cpu.clearFlag(VFlag)
	}

	if sum & NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) and(address uint16) {
	cpu.A &= cpu.memory.Read(address)
	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) aslAcc() {
	carry := cpu.A & NFlag
	cpu.A <<= 1

	if carry > 0 {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) asl(address uint16) {
	carry := cpu.memory.Read(address) & NFlag
	res := cpu.memory.Read(address) << 1
	cpu.memory.Write(address, res)
	
	if carry > 0 {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if res == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if res&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}


func (cpu *Cpu) bcc(address uint16) {
	if cpu.isFlagClear(CFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		cpu.PC = address
	}
}

func (cpu *Cpu) bcs(address uint16) {
	if cpu.isFlagSet(CFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		cpu.PC = address
	}
}

func (cpu *Cpu) beq(address uint16) {
	if cpu.isFlagSet(ZFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		// fmt.Printf("%X\n", address)
		cpu.PC = address
		// fmt.Printf("%X\n", cpu.PC)
		// os.Exit(1)
	}
	// fmt.Printf("-->%X\n", address)
	// os.Exit(1)
}

func (cpu *Cpu) bit(address uint16) {
	a := cpu.A
	m := cpu.memory.Read(address)

	if a&m == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if (m & VFlag) > 0 {
		cpu.setFlag(VFlag)
	} else {
		cpu.clearFlag(VFlag)
	}

	if (m & NFlag) > 0 {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) bmi(address uint16) {
	if cpu.isFlagSet(NFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		cpu.PC = address
	}
}

func (cpu *Cpu) bne(address uint16) {
	if cpu.isFlagClear(ZFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		cpu.PC = address
	}
}

func (cpu *Cpu) bpl(address uint16) {
	if cpu.isFlagClear(NFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		cpu.PC = address
	}
}

func (cpu *Cpu) brk() {
	cpu.irqRequested = true
	cpu.setFlag(BFlag)
}

func (cpu *Cpu) bvc(address uint16) {
	if cpu.isFlagClear(VFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		cpu.PC = address
	}
}

func (cpu *Cpu) bvs(address uint16) {
	if cpu.isFlagSet(VFlag) {
		cpu.cycles += computeCyclesForBranch(cpu.PC, address)
		cpu.PC = address
	}
}

func (cpu *Cpu) clc() {
	cpu.clearFlag(CFlag)
}

func (cpu *Cpu) cld() {
	cpu.clearFlag(DFlag)
}

func (cpu *Cpu) cli() {
	cpu.clearFlag(IFlag)
}

func (cpu *Cpu) clv() {
	cpu.clearFlag(VFlag)
}

func (cpu *Cpu) cmp(address uint16) {
	a := cpu.A
	m := cpu.memory.Read(address)

	if a >= m {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if a == m {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if (a-m)&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) cpx(address uint16) {
	x := cpu.X
	m := cpu.memory.Read(address)

	if x >= m {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if x == m {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if (x-m)&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) cpy(address uint16) {
	y := cpu.Y
	m := cpu.memory.Read(address)

	if y >= m {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if y == m {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if (y-m)&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) dec(address uint16) {
	m := cpu.memory.Read(address)
	res := m - 1
	cpu.memory.Write(address, res)

	if res == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if res&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) dex() {
	cpu.X--

	if cpu.X == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.X&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) dey() {
	cpu.Y--

	if cpu.Y == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.Y&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) eor(address uint16) {
	cpu.A = cpu.A ^ cpu.memory.Read(address)

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) inc(address uint16) {
	m := cpu.memory.Read(address)
	res := m + 1
	cpu.memory.Write(address, res)

	if res == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if res&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) inx() {
	cpu.X++

	if cpu.X == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.X&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) iny() {
	cpu.Y++

	if cpu.Y == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.Y&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) jmp(address uint16) {
	cpu.PC = address
}

func (cpu *Cpu) jsr(address uint16) {
	cpu.pushUint16(cpu.PC - 1)
	cpu.PC = address
}

func (cpu *Cpu) lda(address uint16) {
	cpu.A = cpu.memory.Read(address)

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) ldx(address uint16) {
	cpu.X = cpu.memory.Read(address)
	if cpu.X == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}
	
	if cpu.X & NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}


}

func (cpu *Cpu) ldy(address uint16) {
	cpu.Y = cpu.memory.Read(address)

	if cpu.Y == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.Y&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) lsr(address uint16) {
	//todo?
	m := cpu.memory.Read(address)
	oldBit0 := m & 0x01
	m >>= 1
	cpu.memory.Write(address, m)

	if oldBit0 > 0 {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if m == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if m&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) lsrAcc() {
	//todo
	oldBit0 := cpu.A & 0x01
	cpu.A >>= 1

	if oldBit0 > 0 {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) nop() {
	
}

func (cpu *Cpu) ora(address uint16) {
	cpu.A |= cpu.memory.Read(address)

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) pha() {
	cpu.push(cpu.A)
}

func (cpu *Cpu) php() {
	cpu.push(cpu.P | BFlag)
}

func (cpu *Cpu) pla() {
	cpu.A = cpu.pull()
	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) plp() {
	cpu.P = (cpu.pull() & 0xEF) | 0x20
}

func (cpu *Cpu) rol(address uint16) {
	m := cpu.memory.Read(address)
	
	currentCarry := cpu.getSetFlag(CFlag)
	if m & NFlag == NFlag {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}
	value := (m << 1) | currentCarry

	cpu.memory.Write(address, value)


	if value == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if value&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) rolAcc() {
	currentCarry := cpu.getSetFlag(CFlag)
	
	if cpu.A & NFlag == NFlag {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	value := (cpu.A << 1) | currentCarry
	cpu.A = value

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) ror(address uint16) {
	m := cpu.memory.Read(address)
	
	currentCarry := cpu.getSetFlag(CFlag)
	if m & CFlag == CFlag {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}
	value := (m >> 1) | (currentCarry << 7)

	cpu.memory.Write(address, value)

	
	if value > 0 {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if value&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}

}

func (cpu *Cpu) rorAcc() {
	currentCarry := cpu.getSetFlag(CFlag)
	
	if cpu.A & CFlag == CFlag {
		cpu.setFlag(CFlag)
	} else {
		cpu.clearFlag(CFlag)
	}
	cpu.A = (cpu.A >> 1) | (currentCarry << 7)

	
	
	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) rti() {
	cpu.P = (cpu.pull()&0xEF) | 0x20
	cpu.PC = cpu.pullUint16()
}

func (cpu *Cpu) rts() {
	cpu.PC = cpu.pullUint16() + 1
}

func (cpu *Cpu) sbc(address uint16) {
	a := cpu.A
	m := cpu.memory.Read(address)
	c := cpu.getSetFlag(CFlag)

	cpu.A = a - m - (1 - c)

	if int(a)-int(m)-int(1-c) < 0 {
		cpu.clearFlag(CFlag)
	} else {
		cpu.setFlag(CFlag)
	}

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}


	if (a^m)&NFlag != 0 && (a^cpu.A)&NFlag != 0 {
		cpu.setFlag(VFlag)
	} else {
		cpu.clearFlag(VFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}

}

func (cpu *Cpu) sec() {
	cpu.setFlag(CFlag)
}

func (cpu *Cpu) sed() {
	cpu.setFlag(DFlag)
}

func (cpu *Cpu) sei() {
	cpu.setFlag(IFlag)
}

func (cpu *Cpu) sta(address uint16) {
	cpu.memory.Write(address, cpu.A)
}

func (cpu *Cpu) stx(address uint16) {
	cpu.memory.Write(address, cpu.X)
}

func (cpu *Cpu) sty(address uint16) {
	cpu.memory.Write(address, cpu.Y)
}

func (cpu *Cpu) tax() {
	cpu.X = cpu.A

	if cpu.X == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.X&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) tay() {
	cpu.Y = cpu.A

	if cpu.Y == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.Y&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) tsx() {
	cpu.X = cpu.SP

	if cpu.X == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.X&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) txa() {
	cpu.A = cpu.X

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func (cpu *Cpu) txs() {
	cpu.SP = cpu.X
}

func (cpu *Cpu) tya() {
	cpu.A = cpu.Y

	if cpu.A == 0 {
		cpu.setFlag(ZFlag)
	} else {
		cpu.clearFlag(ZFlag)
	}

	if cpu.A&NFlag == NFlag {
		cpu.setFlag(NFlag)
	} else {
		cpu.clearFlag(NFlag)
	}
}

func MakeNewCpu(nes *NES) *Cpu {
	cpu := Cpu{
		memory: nes,
		P: 0x24,
		SP: 0xFD,
		nmiRequested: false,
		irqRequested: false,
	}
	cpu.PC = cpu.ReadUint16(0xFFFC)

	return &cpu
}

func (cpu *Cpu) promptNMI() {
	cpu.pushUint16(cpu.PC)
	cpu.push(cpu.P)
	cpu.PC = cpu.ReadUint16(0xFFFA)
	cpu.setFlag(IFlag)
	cpu.nmiRequested = false
}

func (cpu *Cpu) promptIRQ() {
	cpu.pushUint16(cpu.PC)
	cpu.push(cpu.P)
	cpu.PC = cpu.ReadUint16(0xFFFE)
	cpu.setFlag(IFlag)
	cpu.irqRequested = false
}

func (cpu *Cpu) run() int{

	if cpu.nmiRequested {
		cpu.promptNMI()
	}
	
	if cpu.irqRequested {
		cpu.promptIRQ()
	}

	opcode := cpu.memory.Read(cpu.PC)
	
	addressingMode := opcodes[opcode].addressingMode
	pageHasCrossed := false
	//currentCpuCycles := cpu.cycles

	var address uint16
	switch addressingMode {
	case imm:
		address = cpu.immediateAddress()
	case zep:
		address = cpu.zeroPageAddress()
	case zpx:
		address = cpu.zeroPageXAddress()
	case zpy:
		address = cpu.zeroPageYAddress()
	case rel:
		address = cpu.relativeAddress()
	case abs:
		address = cpu.absoluteAddress()
	case abx:
		address = cpu.absoluteXAddress()
		pageHasCrossed = isPageCrossed(address, address - uint16(cpu.X))
	case aby:
		address = cpu.absoluteYAddress()
		pageHasCrossed = isPageCrossed(address, address - uint16(cpu.Y))
	case ind:
		address = cpu.indirectAddress()
	case inx:
		address = cpu.indexIndirectAddress()
	case iny:
		address = cpu.indirectIndexedAddress()
		pageHasCrossed = isPageCrossed(address, address - uint16(cpu.Y))
	}

	cpu.PC += uint16(opcodes[opcode].size)
	cpu.cycles += uint64(opcodes[opcode].cycles)
	if pageHasCrossed {
		cpu.cycles += uint64(opcodes[opcode].additionalCycles)
	}

	//updatedCpuCycles := cpu.cycles
	switch opcode {

	case 0x0:
		cpu.brk()
	case 0x1:
		cpu.ora(address)
	case 0x2:
		break
	case 0x3:
		break
	case 0x4:
		cpu.nop()
	case 0x5:
		cpu.ora(address)
	case 0x6:
		cpu.asl(address)
	case 0x7:
		break
	case 0x8:
		cpu.php()
	case 0x9:
		cpu.ora(address)
	case 0xa:
		cpu.aslAcc()
	case 0xb:
		break
	case 0xc:
		cpu.nop()
	case 0xd:
		cpu.ora(address)
	case 0xe:
		cpu.asl(address)
	case 0xf:
		break
	case 0x10:
		cpu.bpl(address)
	case 0x11:
		cpu.ora(address)
	case 0x12:
		break
	case 0x13:
		break
	case 0x14:
		cpu.nop()
	case 0x15:
		cpu.ora(address)
	case 0x16:
		cpu.asl(address)
	case 0x17:
		break
	case 0x18:
		cpu.clc()
	case 0x19:
		cpu.ora(address)
	case 0x1a:
		cpu.nop()
	case 0x1b:
		break
	case 0x1c:
		cpu.nop()
	case 0x1d:
		cpu.ora(address)
	case 0x1e:
		cpu.asl(address)
	case 0x1f:
		break
	case 0x20:
		cpu.jsr(address)
	case 0x21:
		cpu.and(address)
	case 0x22:
		break
	case 0x23:
		break
	case 0x24:
		cpu.bit(address)
	case 0x25:
		cpu.and(address)
	case 0x26:
		cpu.rol(address)
	case 0x27:
		break
	case 0x28:
		cpu.plp()
	case 0x29:
		cpu.and(address)
	case 0x2a:
		cpu.rolAcc()
	case 0x2b:
		break
	case 0x2c:
		cpu.bit(address)
	case 0x2d:
		cpu.and(address)
	case 0x2e:
		cpu.rol(address)
	case 0x2f:
		break
	case 0x30:
		cpu.bmi(address)
	case 0x31:
		cpu.and(address)
	case 0x32:
		break
	case 0x33:
		break
	case 0x34:
		cpu.nop()
	case 0x35:
		cpu.and(address)
	case 0x36:
		cpu.rol(address)
	case 0x37:
		break
	case 0x38:
		cpu.sec()
	case 0x39:
		cpu.and(address)
	case 0x3a:
		cpu.nop()
	case 0x3b:
		break
	case 0x3c:
		cpu.nop()
	case 0x3d:
		cpu.and(address)
	case 0x3e:
		cpu.rol(address)
	case 0x3f:
		break
	case 0x40:
		cpu.rti()
	case 0x41:
		cpu.eor(address)
	case 0x42:
		break
	case 0x43:
		break
	case 0x44:
		cpu.nop()
	case 0x45:
		cpu.eor(address)
	case 0x46:
		cpu.lsr(address)
	case 0x47:
		break
	case 0x48:
		cpu.pha()
	case 0x49:
		cpu.eor(address)
	case 0x4a:
		cpu.lsrAcc()
	case 0x4b:
		break
	case 0x4c:
		cpu.jmp(address)
	case 0x4d:
		cpu.eor(address)
	case 0x4e:
		cpu.lsr(address)
	case 0x4f:
		break
	case 0x50:
		cpu.bvc(address)
	case 0x51:
		cpu.eor(address)
	case 0x52:
		break
	case 0x53:
		break
	case 0x54:
		cpu.nop()
	case 0x55:
		cpu.eor(address)
	case 0x56:
		cpu.lsr(address)
	case 0x57:
		break
	case 0x58:
		cpu.cli()
	case 0x59:
		cpu.eor(address)
	case 0x5a:
		cpu.nop()
	case 0x5b:
		break
	case 0x5c:
		cpu.nop()
	case 0x5d:
		cpu.eor(address)
	case 0x5e:
		cpu.lsr(address)
	case 0x5f:
		break
	case 0x60:
		cpu.rts()
	case 0x61:
		cpu.adc(address)
	case 0x62:
		break
	case 0x63:
		break
	case 0x64:
		cpu.nop()
	case 0x65:
		cpu.adc(address)
	case 0x66:
		cpu.ror(address)
	case 0x67:
		break
	case 0x68:
		cpu.pla()
	case 0x69:
		cpu.adc(address)
	case 0x6a:
		cpu.rorAcc()
	case 0x6b:
		break
	case 0x6c:
		cpu.jmp(address)
	case 0x6d:
		cpu.adc(address)
	case 0x6e:
		cpu.ror(address)
	case 0x6f:
		break
	case 0x70:
		cpu.bvs(address)
	case 0x71:
		cpu.adc(address)
	case 0x72:
		break
	case 0x73:
		break
	case 0x74:
		cpu.nop()
	case 0x75:
		cpu.adc(address)
	case 0x76:
		cpu.ror(address)
	case 0x77:
		break
	case 0x78:
		cpu.sei()
	case 0x79:
		cpu.adc(address)
	case 0x7a:
		cpu.nop()
	case 0x7b:
		break
	case 0x7c:
		cpu.nop()
	case 0x7d:
		cpu.adc(address)
	case 0x7e:
		cpu.ror(address)
	case 0x7f:
		break
	case 0x80:
		cpu.nop()
	case 0x81:
		cpu.sta(address)
	case 0x82:
		cpu.nop()
	case 0x83:
		break
	case 0x84:
		cpu.sty(address)
	case 0x85:
		cpu.sta(address)
	case 0x86:
		cpu.stx(address)
	case 0x87:
		break
	case 0x88:
		cpu.dey()
	case 0x89:
		cpu.nop()
	case 0x8a:
		cpu.txa()
	case 0x8b:
		break
	case 0x8c:
		cpu.sty(address)
	case 0x8d:
		cpu.sta(address)
	case 0x8e:
		cpu.stx(address)
	case 0x8f:
		break
	case 0x90:
		cpu.bcc(address)
	case 0x91:
		cpu.sta(address)
	case 0x92:
		break
	case 0x93:
		break
	case 0x94:
		cpu.sty(address)
	case 0x95:
		cpu.sta(address)
	case 0x96:
		cpu.stx(address)
	case 0x97:
		break
	case 0x98:
		cpu.tya()
	case 0x99:
		cpu.sta(address)
	case 0x9a:
		cpu.txs()
	case 0x9b:
		break
	case 0x9c:
		break
	case 0x9d:
		cpu.sta(address)
	case 0x9e:
		break
	case 0x9f:
		break
	case 0xa0:
		cpu.ldy(address)
	case 0xa1:
		cpu.lda(address)
	case 0xa2:
		cpu.ldx(address)
	case 0xa3:
		break
	case 0xa4:
		cpu.ldy(address)
	case 0xa5:
		cpu.lda(address)
	case 0xa6:
		cpu.ldx(address)
	case 0xa7:
		break
	case 0xa8:
		cpu.tay()
	case 0xa9:
		cpu.lda(address)
	case 0xaa:
		cpu.tax()
	case 0xab:
		break
	case 0xac:
		cpu.ldy(address)
	case 0xad:
		cpu.lda(address)
	case 0xae:
		cpu.ldx(address)
	case 0xaf:
		break
	case 0xb0:
		cpu.bcs(address)
	case 0xb1:
		cpu.lda(address)
	case 0xb2:
		break
	case 0xb3:
		break
	case 0xb4:
		cpu.ldy(address)
	case 0xb5:
		cpu.lda(address)
	case 0xb6:
		cpu.ldx(address)
	case 0xb7:
		break
	case 0xb8:
		cpu.clv()
	case 0xb9:
		cpu.lda(address)
	case 0xba:
		cpu.tsx()
	case 0xbb:
		break
	case 0xbc:
		cpu.ldy(address)
	case 0xbd:
		cpu.lda(address)
	case 0xbe:
		cpu.ldx(address)
	case 0xbf:
		break
	case 0xc0:
		cpu.cpy(address)
	case 0xc1:
		cpu.cmp(address)
	case 0xc2:
		cpu.nop()
	case 0xc3:
		break
	case 0xc4:
		cpu.cpy(address)
	case 0xc5:
		cpu.cmp(address)
	case 0xc6:
		cpu.dec(address)
	case 0xc7:
		break
	case 0xc8:
		cpu.iny()
	case 0xc9:
		cpu.cmp(address)
	case 0xca:
		cpu.dex()
	case 0xcb:
		break
	case 0xcc:
		cpu.cpy(address)
	case 0xcd:
		cpu.cmp(address)
	case 0xce:
		cpu.dec(address)
	case 0xcf:
		break
	case 0xd0:
		cpu.bne(address)
	case 0xd1:
		cpu.cmp(address)
	case 0xd2:
		break
	case 0xd3:
		break
	case 0xd4:
		cpu.nop()
	case 0xd5:
		cpu.cmp(address)
	case 0xd6:
		cpu.dec(address)
	case 0xd7:
		break
	case 0xd8:
		cpu.cld()
	case 0xd9:
		cpu.cmp(address)
	case 0xda:
		cpu.nop()
	case 0xdb:
		break
	case 0xdc:
		cpu.nop()
	case 0xdd:
		cpu.cmp(address)
	case 0xde:
		cpu.dec(address)
	case 0xdf:
		break
	case 0xe0:
		cpu.cpx(address)
	case 0xe1:
		cpu.sbc(address)
	case 0xe2:
		cpu.nop()
	case 0xe3:
		break
	case 0xe4:
		cpu.cpx(address)
	case 0xe5:
		cpu.sbc(address)
	case 0xe6:
		cpu.inc(address)
	case 0xe7:
		break
	case 0xe8:
		cpu.inx()
	case 0xe9:
		cpu.sbc(address)
	case 0xea:
		cpu.nop()
	case 0xeb:
		cpu.sbc(address)
	case 0xec:
		cpu.cpx(address)
	case 0xed:
		cpu.sbc(address)
	case 0xee:
		cpu.inc(address)
	case 0xef:
		break
	case 0xf0:
		cpu.beq(address)
	case 0xf1:
		cpu.sbc(address)
	case 0xf2:
		break
	case 0xf3:
		break
	case 0xf4:
		cpu.nop()
	case 0xf5:
		cpu.sbc(address)
	case 0xf6:
		cpu.inc(address)
	case 0xf7:
		break
	case 0xf8:
		cpu.sed()
	case 0xf9:
		cpu.sbc(address)
	case 0xfa:
		cpu.nop()
	case 0xfb:
		break
	case 0xfc:
		cpu.nop()
	case 0xfd:
		cpu.sbc(address)
	case 0xfe:
		cpu.inc(address)
	case 0xff:
		break
	}

	return int(opcodes[opcode].cycles)
}

func computeCyclesForBranch(pc uint16, addr uint16) uint64{
	if isPageCrossed(pc, addr) {
		return 2
	} else {
		return 1
	}
}

func isPageCrossed(a uint16, b uint16) bool {
	//http://forums.nesdev.com/viewtopic.php?f=3&t=365
	return a&0xFF00 != b&0xFF00
}

func (cpu *Cpu) setFlag(flag byte) {
	cpu.P |= flag
}

func (cpu *Cpu) clearFlag(flag byte) {
	cpu.P &= (flag^0xFF)
}

func (cpu *Cpu) isFlagSet(flag byte) bool {
	return cpu.P&flag > 0
}
func (cpu *Cpu) isFlagClear(flag byte) bool {
	return !cpu.isFlagSet(flag)
}

func (cpu *Cpu) getSetFlag(flag byte) byte {
	if cpu.isFlagSet(flag) {
		return 1
	}
	return 0
}

func (cpu *Cpu) push(value byte) {
	cpu.memory.Write(uint16(cpu.SP)|0x100, value)
	cpu.SP--
}

func (cpu *Cpu) pushUint16(value uint16) {
	cpu.push(byte(value >> 8))
	cpu.push(byte(value & 0xFF))
}

func (cpu *Cpu) pull() byte {
	cpu.SP++
	r := cpu.memory.Read(uint16(cpu.SP) | 0x100)
	return r
}

func (cpu *Cpu) pullUint16() uint16 {
	low := uint16(cpu.pull())
	high := uint16(cpu.pull())
	return high<<8 | low
}
