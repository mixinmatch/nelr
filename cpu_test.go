package main

import(
	"testing"
	"os"
	"bufio"
	"log"
	"strings"
	"regexp"
	"fmt"
)

func TestCpuOpcodes(t *testing.T) {
	romPath := "./roms/test/nestest.nes"
	cart := LoadRom(romPath)
	nes := NES{}
	nes.cart = &cart
	nes.mapper = MakeNewMapper(&nes)
	nes.ppu = MakeNewPPU(&nes)
	nes.cpu = MakeNewCpu(&nes)

	//nestest.nes starts at $C000
	nes.cpu.PC = 0xC000
	nes.cpu.cycles = 7

	expectedLogPath := "./roms/test/nestest.log"
	file, err := os.Open(expectedLogPath)
	defer file.Close()	
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	lineCount := 0 
	for scanner.Scan() {
		lineCount++

		//Do not test first-encountered illegal opcode at L5261
		if lineCount == 5261 { 
			break
		}
		
		line := scanner.Text()
		s := strings.Split(line, " ")
		c := make([]string, 0)
		for _, e := range s {
			if e != "" {
				c = append(c, e)
			}
		}

		re := regexp.MustCompile(`(A:\w\d)|(A:\d\w)|(A:\w\w)|(A:\d\d)`)
		var aIdx int
		for i, _ := range c {
			if re.MatchString(c[i]) {
				aIdx = i
			}
		}
		
		expectedPC := c[0]
		expectedA := c[aIdx]
		expectedX := c[aIdx + 1]
		expectedY := c[aIdx + 2]
		expectedP := c[aIdx + 3]
		expectedSP := c[aIdx + 4]
		// expectedScanline := 0
		// expectedPPUCycle := 0
		expectedCycle := c[len(c)-1]

		cpu := nes.cpu
		actualPC := fmt.Sprintf("%04X", cpu.PC)
		actualA := fmt.Sprintf("A:%02X", cpu.A)
		actualX := fmt.Sprintf("X:%02X", cpu.X)
		actualY := fmt.Sprintf("Y:%02X", cpu.Y)
		actualP := fmt.Sprintf("P:%X", cpu.P)
		actualSP := fmt.Sprintf("SP:%X", cpu.SP)
		// actualScanline := fmt.Sprintf()
		// actualPPUCycle := fmt.Sprintf()
		actualCycle := fmt.Sprintf("CYC:%v", cpu.cycles)

		printDiff := func(whichRegister string) {
			
			t.Errorf("%v @Line:%v \n Expected: %v %v %v %v %v %v %v\n      Got: %04v %v %v %v %v %v %v\n", whichRegister, lineCount, expectedPC, expectedA, expectedX,
				expectedY, expectedP, expectedSP, expectedCycle , actualPC, actualA, actualX, actualY, actualP, actualSP, actualCycle)
		}
		
		if expectedPC != actualPC {			
			printDiff("PC")
			break
		}
		if expectedA != actualA {
			printDiff("A")
			break
		}
		if expectedX != actualX {
			printDiff("X")
			break	
		}
		if expectedY != actualY {
			printDiff("Y")
			break
		}
		if expectedP != actualP {
			printDiff("P")
			break
		}
		if expectedSP != actualSP {
			printDiff("SP")
			break
		}
		if expectedCycle != actualCycle {
			printDiff("CYC")
			break
		}
		
		nes.cpu.run()

	}

}

// func TestCpuWrite(t *testing.T) {
// 	romPath := "./roms/test/cpu_dummy_writes_ppumem.nes"
// 	cart := LoadRom(romPath)
// 	ram := make([]byte, 0xFFFF+1)
// 	nes := NES{ram: ram}
// 	nes.cart = &cart
// 	nes.mapper = MakeNewMapper(&nes)
// 	nes.ppu = MakeNewPPU(&nes)
// 	nes.cpu = MakeNewCpu(&nes)

// 	for i := 0; i < 100000; i++{
// 	nes.cpu.run()
// 	}
// }
