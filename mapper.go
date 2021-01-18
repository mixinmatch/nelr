package main

import (
	"log"
)

type Mapper interface {
	Read(addr uint16) byte
	Write(addr uint16, value byte)
}

func MakeNewMapper(nes *NES) Mapper {
	mapperId := nes.cart.getMapperId()

	switch mapperId {

	case 0:
		return MakeNewMapper0(nes)

	default:
		log.Fatalf("mapper %v not supported\n", mapperId)
	}

	return MakeNewMapper0(nes)
	
}
