package main

import (
_	"fmt"
	"log"
_	"os"
_	"runtime/debug"
)

type PPU struct {
	nes *NES

	ctrl    byte
	mask    byte
	status  byte
	oamaddr byte
	oamdata byte
	scroll  byte
	addr    byte
	data    byte
	oamdma  byte

	v uint16
	t uint16
	x byte
	w byte

	lastRegisterWrite byte

	//Sprites
	oam [256]byte

	spritePosition        [8]byte
	spritePatterns        [8]uint32
	spriteIds       [8]int
	spritePriority [8]byte
	
	spriteInScanlineCount int

	vram        [2048]byte //nametables
	paletteInfo [32]byte
	palette     [64]uint32

	cycles   int
	scanline int

	nametableLatch byte
	attributeLatch byte
	patternLowLatch byte
	patternHighLatch byte
	
	backgroundTile uint64
}

func (ppu *PPU) Run() {
	isPreScanline := ppu.scanline == 261
	isRenderingScanline := ppu.scanline <= 239
	isVerticalBlank := ppu.scanline == 241

	//fmt.Printf("SC:%v CYC:%v $V:%X $T:%X\n",ppu.scanline, ppu.cycles, ppu.v, ppu.t)

	if isPreScanline {
		ppu.PreRenderScanline()
	}

	if isRenderingScanline {
		ppu.RenderVisibleScanline()
	}

	if isVerticalBlank {
		if ppu.cycles == 1 {
			ppu.setVBlank()
		}
	}

	ppu.cycles++
	if ppu.cycles == 341 {
		ppu.scanline++
		if ppu.scanline == 262 {
			drawFrame()
			ppu.scanline = 0
		}
		ppu.cycles = 0
	}
}

func (ppu *PPU) PreRenderScanline() {
	if ppu.cycles == 1 {
		ppu.clearVBlank()
		ppu.clearSprite0Hit()
		ppu.clearSpriteOverflow()
	}

	if ppu.cycles >= 280 && ppu.cycles <= 304 {
		if ppu.IsRenderingEnabled() {
			ppu.CopyVertical()
		}
	}

	if ppu.cycles == 257 {
		ppu.spriteInScanlineCount = 0
	}
	
	ppu.RenderVisibleScanline()
}

func (ppu *PPU) RenderVisibleScanline() {

	if !ppu.IsRenderingEnabled() {
		return
	}

	//fmt.Printf("-v:%X \n", ppu.v)
	if ppu.cycles >= 1 && ppu.cycles <= 256 {
		ppu.renderPixel()
	}
	
	if isFetchTime := ppu.cycles >= 1 && ppu.cycles <= 256 || ppu.cycles >= 321 && ppu.cycles <= 336; isFetchTime {
		ppu.backgroundTile <<= 4
		switch ppu.cycles%8 {
		case 1: ppu.nametableLatch = ppu.fetchNametableByte()
		case 3: ppu.attributeLatch = ppu.fetchAttributetableByte()
		case 5: ppu.patternLowLatch = ppu.fetchPatternTableLowByte()
		case 7: ppu.patternHighLatch = ppu.fetchPatternTableHighByte()
		case 0: ppu.backgroundTile |= uint64(ppu.makeBgTile())
			ppu.incrementHorizontal()
		}
	}
	//fmt.Printf("v:%X \n", ppu.v)

	if ppu.cycles == 257 {
		ppu.doSpriteEvaluation()
	}
	//fmt.Printf("v:%X \n", ppu.v)


	if ppu.cycles == 256 {
		ppu.incrementVertical()
	}
	//fmt.Printf("v:%X \n", ppu.v)

	if ppu.cycles == 257 {
		ppu.CopyHorizontal()
	}
	//fmt.Printf("v:%X \n", ppu.v)
}

func (ppu *PPU) makeBgTile() uint32 {
	var tile uint32
	for i := 0; i < 8; i++ {
		// Each bit in bitplane represent a single pixel
		a := ppu.attributeLatch << 2
		p0 := (ppu.patternLowLatch&0x80)>>7
		p1 := (ppu.patternHighLatch&0x80)>>6
		
		ppu.patternLowLatch <<= 1
		ppu.patternHighLatch <<= 1
		//One pixel is rep by 4 bits, one tile - 8 bit wide - is 32 bits
		tile <<= 4
		tile |= uint32(a | p0 | p1)
	}

	return tile
}
func (ppu *PPU) fetchNametableByte() byte {
	addr := 0x2000 | (ppu.v & 0x0FFF)
	nametableByte := ppu.Read(addr)
	//fmt.Printf("V:$%X addr:$%X nametablebyte:$%X \n", ppu.v, addr /* asMirroredAddress(addr, ppu.nes.cart.getMirroringId()),*/, nametableByte)

	return nametableByte
}

func (ppu *PPU) fetchAttributetableByte() byte {
	v := ppu.v
	tableAddr := 0x23C0 | (v & 0x0C00) | ((v >> 4) & 0x38) | ((v >> 2) & 0x07)

	//deduce how much to shift from nametable address to pick correct palette
	//https://forums.nesdev.com/viewtopic.php?f=3&t=14795
	//http://wiki.nesdev.com/w/index.php/PPU_attribute_tables
	//has 4 "blocks" of 2x2 tiles

	s := (v>>4)&4 | v&2
	attributeByte := (ppu.Read(tableAddr) >> s)&3
	
	//fmt.Printf("addr: $%X attributebyte: %X \n", tableAddr, attributeByte)


	return attributeByte
}

func (ppu *PPU) fetchPatternTableLowByte() byte {
	patternTable := 0x1000 * uint16((ppu.ctrl>>4)&1)
	fineYScroll := (ppu.v>>12) & 0x07
	//fmt.Printf("V:$%X, PT:$%X, NT:$%X, YSC:$%X\n", ppu.v, patternTable, uint16(nametableTile)*16, fineYScroll)

	nametableTile := ppu.nametableLatch
	ptTileLowAddr := patternTable + uint16(nametableTile)*16 + fineYScroll
	ptTile := ppu.Read(ptTileLowAddr)
	//fmt.Printf("addr:$%X tileLow: %X \n",ptTileLowAddr, ptTile)

	return ptTile
}

func (ppu *PPU) fetchPatternTableHighByte() byte {
	patternTable := 0x1000 * uint16((ppu.ctrl>>4)&1)
	fineYScroll := (ppu.v>>12) & 0x07
	nametableTile := ppu.nametableLatch
	ptTileLowAddr := patternTable + uint16(nametableTile)*16 + fineYScroll
	ptTile := ppu.Read(ptTileLowAddr + 8)

	//fmt.Printf("addr:$%X tileHigh: %X \n",ptTileLowAddr+8, ptTile)

	return ptTile
}

func (ppu *PPU) IsRenderingEnabled() bool {
	return ppu.IsBackgroundEnabled() || ppu.IsSpriteEnabled()
}

func (ppu *PPU) IsBackgroundEnabled() bool {
	return (ppu.mask>>3)&1 == 1
}

func (ppu *PPU) IsSpriteEnabled() bool {
	return (ppu.mask>>4)&1 == 1
}

func (ppu *PPU) renderPixel() {
	x := int(ppu.cycles - 1)
	y := int(ppu.scanline)

	bgPixel := ppu.getBackgroundPixel()
	spritePixel, i :=  ppu.getSpritePixel()

	//pixel multiplexer computation
	bgIsOpaque := bgPixel%4 != 0
	spriteIsOpaque := spritePixel%4 != 0
	bgIsTransparent := !bgIsOpaque
	spriteIsTransparent := !spriteIsOpaque
	
	var color byte
	if bgIsTransparent && spriteIsTransparent {
		color = 0
	} else if spriteIsOpaque && bgIsTransparent {
		color = spritePixel | 0x10
	} else if spriteIsTransparent && bgIsOpaque {
		color = bgPixel
	} else { //Both are opaque
		if ppu.spriteIds[i] == 0 && x < 255 {
			ppu.setSprite0Hit()
		}

		if ppu.spritePriority[i] == 0 {
			color = spritePixel | 0x10
		} else {
			color = bgPixel
		}
	}
	
	paletteIndex := ppu.ReadPalette(uint16(color)) % 64
	pixelColor := ppu.palette[paletteIndex]

	var a byte = 0xFF //always opaque
	r := byte(pixelColor>>16) & 0xFF
	g := byte(pixelColor>>8) & 0xFF
	b := byte(pixelColor>>0) & 0xFF

	if x >= 0 && x <= 256 && y < windowHeight { //only render 240 scanline
		renderBuffer[(y*windowWidth+x)*4+0] = b
		renderBuffer[(y*windowWidth+x)*4+1] = g
		renderBuffer[(y*windowWidth+x)*4+2] = r
		renderBuffer[(y*windowWidth+x)*4+3] = a
	}	
}

func (ppu *PPU) getBackgroundPixel() byte {
	if !ppu.IsBackgroundEnabled() {
		return 0
	}
	
	tile := ppu.backgroundTile >> 32 //tile shifted during fetch with << 4
	tile = (tile >> ((7 - ppu.x) * 4)) & 0xF //pick tile, tile is 4 bits

	return byte(tile)
}

func (ppu *PPU) getSpritePixel() (byte, int)  {
	if !ppu.IsSpriteEnabled() {
		return 0,0
	}

	isWithinRange := func(x byte, currentDot int) bool {
		return int(x)<=currentDot && currentDot<int(x+8)
	}
	currentDot := ppu.cycles-1
	for i := 0; i < ppu.spriteInScanlineCount; i++ {
		if isWithinRange(ppu.spritePosition[i], currentDot) {
			shift := 4*(7 - (currentDot - int(ppu.spritePosition[i])))
			color := byte((ppu.spritePatterns[i]>>shift)&0xF)
			sid := i

			if color%4 != 0 {
				return color, sid
			}

		}
	}

	return 0,0
}
func (ppu *PPU) doSpriteEvaluation() {
	var spriteHeight int
	if (ppu.ctrl>>5)&1 == 0 {
		spriteHeight = 8
	} else {
		spriteHeight = 16
	}

	spriteCount := 0
	currentScanline := ppu.scanline
	for i:=0; i<64; i++ {
		y := ppu.oam[i*4+0]
		t := ppu.oam[i*4+1]
		a := ppu.oam[i*4+2]
		x := ppu.oam[i*4+3]

		offset := currentScanline-int(y)
		spriteOnThisScanline := offset < spriteHeight && offset >= 0
		if !spriteOnThisScanline {
			continue
		}

		if spriteCount < 8 {
			ppu.spritePosition[spriteCount] = x
			ppu.spritePriority[spriteCount] = (a>>5)&1
			ppu.spritePatterns[spriteCount] = ppu.getSpritePatterns(a, t, offset)
			ppu.spriteIds[spriteCount] = i
		}

		spriteCount++
	}

	if spriteCount > 8 {
		spriteCount = 8
		ppu.setSpriteOverflow()
	}
	ppu.spriteInScanlineCount = spriteCount
}

func (ppu *PPU) getSpritePatterns(a byte, t byte, row int) uint32 {
	var addr uint16
	if (ppu.ctrl>>5)&1 == 0 { //sprite height is 8
		if isSpriteVerticalFlip(a) {
			row = 7 - row
		}
		spriteBaseTableAddr := 0x1000 * uint16((ppu.ctrl>>3)&1)
		addr = spriteBaseTableAddr + uint16(t)*16 + uint16(row)
	} else { //sprite height is 16
		if isSpriteVerticalFlip(a) {
			row = 15 - row
		}

		spriteBaseTableAddr := 0x1000 * uint16(t&1)
		t&=0xFE
		if row > 7 {
			t++
			row -= 8
		}
		
		addr = spriteBaseTableAddr + uint16(t)*16 + uint16(row)
	}

	palette := (a&3) << 2
	patternLow := ppu.Read(addr)
	patternHigh := ppu.Read(addr+8)

	//Form sprite
	var p0, p1 byte
	var completeSprite uint32
	for i:=0; i < 8; i++ {
		if isSpriteHorizontalFlip(a) {
			p0 = (patternLow&1)<<0 
			p1 = (patternHigh&1)<<1
			patternLow >>= 1
			patternHigh >>= 1
		} else {
			p0 = (patternLow>>7)&1
			p1 = (patternHigh>>6)&2
			patternLow <<= 1
			patternHigh <<= 1
		}
		completeSprite <<= 4
		completeSprite |= uint32(palette | p1 | p0)
	}
	return completeSprite
}

func (ppu *PPU) CopyVertical() {
	ppu.v = (ppu.v & 0x841F) | (ppu.t & 0x7BE0)
}

func (ppu *PPU) CopyHorizontal() {
	ppu.v = (ppu.v & 0xFBE0) | (ppu.t & 0x841F)
}

func (ppu *PPU) incrementHorizontal() {
	if (ppu.v & 0x001F) == 31 {
		ppu.v &= 0xFFE0
		ppu.v ^= 0x0400
	} else {
		ppu.v += 1
	}
}

func (ppu *PPU) incrementVertical() {

	if (ppu.v & 0x7000) != 0x7000 {
		ppu.v += 0x1000
	} else {
		ppu.v &= 0x8fff
		var y uint16 = (ppu.v&0x03E0) >> 5

		if y == 29 {
			y = 0
			ppu.v ^= 0x0800
		} else if y == 31 {
			y = 0
		} else {
			y += 1
		}
		ppu.v = (ppu.v&0xFC1F) | (y<<5)
	}
}

func isSpriteHorizontalFlip(attribute byte) bool {
	return (attribute>>6)&1 == 1
}

func isSpriteVerticalFlip(attribute byte) bool {
	return (attribute>>7)&1 == 1
}

func (ppu *PPU) setVBlank() {
	//nmiOccured
	ppu.status|=(1<<7)


	//nmiOutput
	if (ppu.ctrl>>7)&1 == 1 {
		ppu.nes.cpu.nmiRequested = true
	}
}

func (ppu *PPU) clearVBlank() {
	//nmiOccured false
	ppu.status &= 0x7F
}

func (ppu *PPU) setSpriteOverflow() {
	ppu.status |= (1 << 5)
}

func (ppu *PPU) clearSpriteOverflow() {
	mask := byte(0xFF ^ (1 << 5))
	ppu.status &= mask
}

func (ppu *PPU) setSprite0Hit() {
	ppu.status |= (1 << 6)
}

func (ppu *PPU) clearSprite0Hit() {
	mask := byte(0xFF ^ (1 << 6))
	ppu.status &= mask
}

func MakeNewPPU(nes *NES) *PPU {
	p := [64]uint32{
		0x7C7C7C, 0x0000FC, 0x0000BC, 0x4428BC, 0x940084, 0xA80020, 0xA81000, 0x881400,
		0x503000, 0x007800, 0x006800, 0x005800, 0x004058, 0x000000, 0x000000, 0x000000,
		0xBCBCBC, 0x0078F8, 0x0058F8, 0x6844FC, 0xD800CC, 0xE40058, 0xF83800, 0xE45C10,
		0xAC7C00, 0x00B800, 0x00A800, 0x00A844, 0x008888, 0x000000, 0x000000, 0x000000,
		0xF8F8F8, 0x3CBCFC, 0x6888FC, 0x9878F8, 0xF878F8, 0xF85898, 0xF87858, 0xFCA044,
		0xF8B800, 0xB8F818, 0x58D854, 0x58F898, 0x00E8D8, 0x787878, 0x000000, 0x000000,
		0xFCFCFC, 0xA4E4FC, 0xB8B8F8, 0xD8B8F8, 0xF8B8F8, 0xF8A4C0, 0xF0D0B0, 0xFCE0A8,
		0xF8D878, 0xD8F878, 0xB8F8B8, 0xB8F8D8, 0x00FCFC, 0xF8D8F8, 0x000000, 0x000000,
	}
	ppu := PPU{
		nes:     nes,
		palette: p,
	}
	ppu.Reset()
	return &ppu
}

func (ppu *PPU) Reset() {
	ppu.scanline = 0
	ppu.cycles = 340

	ppu.WriteCtrl(0x00)
	ppu.WriteMask(0x00)
	ppu.WriteOamAddr(0x00)
	ppu.WriteScroll(0x00)
}

func (ppu *PPU) WriteRegisters(addr uint16, data byte) {
	ppu.lastRegisterWrite = data

	//log.Printf("%x\n", addr)
	switch {
	case addr == 0x2000:
		ppu.WriteCtrl(data)
	case addr == 0x2001:
		ppu.WriteMask(data)
	case addr == 0x2003:
		ppu.WriteOamAddr(data)
	case addr == 0x2004:
		ppu.WriteOamData(data)
	case addr == 0x2005:
		ppu.WriteScroll(data)
	case addr == 0x2006:
		ppu.WriteAddress(data)
	case addr == 0x2007:
		ppu.WriteData(data)
	case addr == 0x4014:
		ppu.WriteOamDma(data)
	default:
		log.Fatalf("$%x is invalid", addr)
	}
}

func (ppu *PPU) WriteCtrl(data byte) {
	ppu.ctrl = data
	ppu.t = (ppu.t & 0xF3FF) | ((uint16(data) & 0x3) << 10)
}

func (ppu *PPU) WriteMask(data byte) {
	ppu.mask = data
}

func (ppu *PPU) WriteOamAddr(data byte) {
	ppu.oamaddr = data
}

func (ppu *PPU) WriteOamData(data byte) {
	ppu.oam[ppu.oamaddr] = data
	ppu.oamaddr++
}

func (ppu *PPU) WriteScroll(data byte) {
	if ppu.w == 0 {
		ppu.t = (ppu.t & 0xFFE0) | (uint16(data) >> 3) 
		ppu.x = data & 0x7
		ppu.w = 1
	} else {
		t := ppu.t
		ppu.t = (t & 0x8FFF) | (t & 0xFC1F) | (uint16(data)&0x7)<<12 | (uint16(data) & 0xF8)<<2
		ppu.w = 0
	}
}

func (ppu *PPU) WriteAddress(data byte) {
	if ppu.w == 0 {
		//t: .FEDCBA ........ = d: ..FEDCBA
		//t: X...... ........ = 0
		ppu.t = (ppu.t & 0x80FF) | ((uint16(data) & 0x3F) << 8)
		ppu.w = 1
	} else {
		ppu.t = (ppu.t & 0xFF00) | uint16(data)
		ppu.v = ppu.t
		ppu.w = 0
	}
}
func (ppu *PPU) WriteData(data byte) {
	ppu.Write(ppu.v, data)
	ppu.IncrementV()
}

func (ppu *PPU) WriteOamDma(data byte) {
	addr := uint16(data) << 8 //Data taken from $XX00 to $XXFF from CPU memory 
	for i := 0; i < 256; i++ {
		ppu.oam[ppu.oamaddr] = ppu.nes.cpu.memory.Read(addr)
		ppu.oamaddr++
		addr++
	}

	ppu.nes.cpu.suspendCycles += 513
	if ppu.nes.cpu.cycles%2 == 1 {
		ppu.nes.cpu.suspendCycles++
	}
}

func (ppu *PPU) ReadRegisters(addr uint16) byte {
	switch {
	case addr == 0x2002:
		return ppu.ReadStatus()
	case addr == 0x2004:
		return ppu.ReadOamData()
	case addr == 0x2007:
		return ppu.ReadData()
	default:
		log.Fatalf("$%x is invalid address", addr)
	}
	return 0
}

func (ppu *PPU) ReadStatus() byte {
	s := ppu.status & (1<<7|1<<6|1<<5)
	s |= ppu.lastRegisterWrite&0x1F
	
	//nmiOccurred false
	ppu.status &= 0x7F

	ppu.w = 0

	return s
}

func (ppu *PPU) ReadOamData() byte {
	return ppu.oam[ppu.oamaddr]
}

func (ppu *PPU) ReadData() byte {
	data := ppu.Read(ppu.v)
	ppu.IncrementV()

	return data
}

func (ppu *PPU) IncrementV() {
	vIncrement := ppu.ctrl & (1 << 2)
	if vIncrement == 0 {
		ppu.v += 1
	} else {
		ppu.v += 32
	}
}

func (ppu *PPU) WritePalette(addr uint16, value byte) {
	ppu.paletteInfo[addr] = value
}

func (ppu *PPU) ReadPalette(addr uint16) byte {
	return ppu.paletteInfo[addr]
}
