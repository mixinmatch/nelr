package main

import (
	"log"
)

type Mapper0 struct {
	nes *NES
}
//https://wiki.nesdev.com/w/index.php/INES_Mapper_000
func (mapper Mapper0) Read(addr uint16) byte {

	switch {
	case addr < 0x2000:
		return mapper.nes.cart.chr[addr]
	case addr >= 0x8000:
		a := (addr-0x8000) % (0x4000*uint16(mapper.nes.cart.header.PrgRomSize))
		return mapper.nes.cart.prg[a]	
	default:
		log.Fatalf("$%x is invalid", addr)
	}
	
	return 0
}

func (mapper Mapper0) Write(addr uint16, value byte) {
	switch {
	//Rom is read-only.
	case addr >= 0x8000:
		
	default:
		log.Fatalf("$%x is invalid", addr)

	}	
}

func MakeNewMapper0(nes *NES) Mapper0 {
	return Mapper0{nes: nes}
}
