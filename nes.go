package main


import(
_	"fmt"

)
type NES struct {
	cpu *Cpu
	ram [0xFFFF+1]byte
	ppu *PPU
	controller *GameController
	mapper Mapper
	cart *Cartridge
}

func MakeNewNES(cartridge *Cartridge) NES {
	nes := NES{
		cart: cartridge,		
	}
	nes.cart = cartridge
	nes.mapper = MakeNewMapper(&nes)
	nes.ppu = MakeNewPPU(&nes)
	nes.cpu = MakeNewCpu(&nes)
	nes.controller = MakeNewGameController()

	return nes
}

func (nes *NES) Run() {
	//fmt.Printf("-nes.ppu.t: %v\n", nes.ppu.t)
	cycles := nes.cpu.run()
	//fmt.Printf("nes.ppu.t: %v\n", nes.ppu.t)
	for i:=0; i<3*cycles; i++{
		nes.ppu.Run()
	}
}
