package main

import (
	"io"
	"os"
	"log"
	"encoding/binary"
)

const iNESMagicNumber = 0x1A53454E

type INESHeader struct {
	MagicNumber uint32
	PrgRomSize  byte
	ChrRomSize  byte
	Flag6       byte
	Flag7       byte
	PrgRamSize  byte
	ExtraFlags  [7]byte //all 0s
}

type Cartridge struct {
	header  INESHeader
	trainer []byte
	prg []byte
	chr []byte
	wram [0x2000]byte //Not in cartridge but in NROM
}

func LoadRom(path string) Cartridge {
	var err error
	var rom *os.File
	rom, err = os.Open(path)
	checkError(err)
	defer rom.Close()

	header := INESHeader{}
	err = binary.Read(rom, binary.LittleEndian, &header)
	checkError(err)

	if header.MagicNumber != iNESMagicNumber {
		log.Fatal("Cannot parse non INES roms")
	}

	var trainerSize uint32
	if header.Flag6&0x4 > 0 {
		trainerSize = 512
	} else {
		trainerSize = 0
	}
	trainer := readNextNBytes(rom, trainerSize)

	prgSize := 16384*uint32(header.PrgRomSize)
	prg := readNextNBytes(rom, prgSize)

	chrSize := 8192*uint32(header.ChrRomSize)
	chr := readNextNBytes(rom, chrSize)

	cartridge := Cartridge{
		header: header,
		trainer: trainer,
		prg: prg,
		chr: chr}
	return cartridge	
}

func readNextNBytes(rom *os.File, size uint32) []byte {
	block := make([]byte, size)
	_ , err = io.ReadFull(rom, block)
	checkError(err)

	return block
}

func (cart *Cartridge) getMapperId() byte {
	return (cart.header.Flag6 & 0xf0) >> 4
}

func (cart *Cartridge) getMirroringId() byte {
	return cart.header.Flag6 & 0x01
}
