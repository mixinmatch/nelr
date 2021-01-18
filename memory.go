package main

import (
	"log"
)

type Memory interface {
	Read(addr uint16) byte
	Write(addr uint16, value byte)
}

func (cpu *Cpu) ReadUint16(addr uint16) uint16 {
	return uint16(cpu.memory.Read(addr)) | uint16(uint16(cpu.memory.Read(addr+1))<<8)
}

//cpu memory map
func (nes *NES) Read(addr uint16) byte {

	switch {
	case addr < 0x2000:
		return nes.ram[addr%0x0800]
	case addr < 0x4000:
		a := addr%8 + 0x2000
		return nes.ppu.ReadRegisters(a)
	case addr == 0x4015:
		return 0 //TODO APU
	case addr == 0x4016:
		return nes.controller.Read()
	case addr == 0x4017:
		return 0 //TODO joy stick 2
	case addr >= 0x4000 && addr < 0x6000:
		return 0 //TODO APU and IO Registers
	case addr >= 0x6000 && addr < 0x8000:
		a := addr - 0x6000
		return nes.cart.wram[a]
	case addr >= 8000:
		return nes.mapper.Read(addr)
	default:
		log.Fatalf("$%x is invalid", addr)
	}
	return 0
}

func (nes *NES) Write(addr uint16, content byte) {
	switch {
	case addr < 0x2000:
		nes.ram[addr % 0x0800] = content
	case addr < 0x4000:
		a := addr%8 + 0x2000
		nes.ppu.WriteRegisters(a, content)
	case addr == 0x4014:
		nes.ppu.lastRegisterWrite = content
		nes.ppu.WriteOamDma(content)
	case addr == 0x4015:
		//TODO APU
	case addr == 0x4016:
		nes.controller.Write(content)
	case addr == 0x4017:
		//TODO joy stick 2
	case addr >= 0x4000 && addr < 0x6000:
		//TODO APU and IO Registers
	case addr >= 0x6000 && addr < 0x8000:
		a := addr - 0x6000
		nes.cart.wram[a] = content
	case addr >= 0x8000:
		nes.mapper.Write(addr, content)
	default:
		log.Fatalf("$%x is invalid\n", addr)
	}
}

//ppu memory map
//https://wiki.nesdev.com/w/index.php/PPU_memory_map
func (ppu *PPU) Read(addr uint16) byte {
	addr %= 0x4000
	switch {
	case addr < 0x2000:
		return ppu.nes.mapper.Read(addr)
	case addr < 0x3F00: //Maps from $2000-$3EFF

		if addr >= 0x3000 {
			addr -= 0x1000
		}
		
		mirroringMode := ppu.nes.cart.getMirroringId()
		mirroredAddr := asMirroredAddress(addr, mirroringMode)
		return ppu.vram[mirroredAddr]
	case addr < 0x4000:
		return ppu.ReadPalette(addr%32)
	default:
		log.Fatalf("$%x is invalid", addr)
	}
	return 0
}

func (ppu *PPU) Write(addr uint16, value byte) {
	addr %= 0x4000
	switch {
	case addr < 0x3F00: //Maps from $2000-$3EFF
		if addr >= 0x3000 {
			addr -= 0x1000
		}
		
		mirroringMode := ppu.nes.cart.getMirroringId()
		mirroredAddr := asMirroredAddress(addr, mirroringMode)
		ppu.vram[mirroredAddr] = value //only two name tables
	case addr < 0x4000:
		ppu.WritePalette(addr%32, value)
	default:
		log.Fatalf("$%x is invalid", addr)
	}
}

//ppu memory map addr->vram
func asMirroredAddress(addr uint16, mirrorId byte) uint16 {
	isHorizontalMirroring := mirrorId == 0
	if isHorizontalMirroring {
		var id []uint16 = []uint16{0,0,1,1}
		a := (addr-0x2000)%0x1000
		return id[a/0x400]*0x400 + a%0x400
	} else {
		return addr%0x800
	}	
}

