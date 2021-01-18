package main
//https://forums.nesdev.com/viewtopic.php?f=3&t=10297
//https://github.com/kingcons/famiclom/blob/master/docs/nes.txt

import (
	"log"
	"github.com/veandco/go-sdl2/sdl"
	"os"
	"fmt"
)

const (
	windowHeight = 240
	windowWidth = 256
	argbBytes = 4
)

var renderBuffer [windowWidth * windowWidth * argbBytes]byte
var texture *sdl.Texture
var err error
var renderer *sdl.Renderer
var window *sdl.Window

//https://wiki.libsdl.org/MigrationGuide
func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: nelr <rom.nes>")
		os.Exit(1)
	}
	
	log.SetFlags(log.Lshortfile)
	cart := LoadRom(os.Args[1])
	nes := MakeNewNES(&cart)
	nes.ppu.Reset()

	sdl.Init(sdl.INIT_EVERYTHING)
	window, renderer, err = sdl.CreateWindowAndRenderer(windowWidth, windowHeight, sdl.WINDOW_RESIZABLE)
	checkError(err)
	texture, err = renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888,	sdl.TEXTUREACCESS_STREAMING, windowWidth, windowHeight)
	checkError(err)
	
	var isRunning = true
	for isRunning {
		//log.Println(nes.ppu.t)
		nes.Run()
		
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				isRunning = false
				texture.Destroy()
				renderer.Destroy()
				window.Destroy()
				sdl.Quit()

			case *sdl.KeyboardEvent:
				keyIsReleased := t.Type == sdl.KEYUP
				keyIsPressed := t.Type == sdl.KEYDOWN
				keyScancode := t.Keysym.Scancode
				// log.Printf("keyPressed:%v keyReleased:%v scancode:%v \n", keyIsPressed, keyIsReleased,  keyScancode)
				if keyIsPressed {
					nes.controllerButtonPressed(keyScancode)
				}
				if keyIsReleased {
					nes.controllerButtonReleased(keyScancode)
				}

			}
		}
	}
}

func (nes *NES) controllerButtonPressed(scancode sdl.Scancode) {
	if scancode == sdl.SCANCODE_A {
		nes.controller.pressButton(controllerButtonA)
	} else if scancode == sdl.SCANCODE_B {
		nes.controller.pressButton(controllerButtonB)
	} else if scancode == sdl.SCANCODE_Z {
		nes.controller.pressButton(controllerButtonSelect)					
	} else if scancode == sdl.SCANCODE_X {
		nes.controller.pressButton(controllerButtonStart)
	} else if scancode == sdl.SCANCODE_UP {
		nes.controller.pressButton(controllerButtonUp)
	} else if scancode == sdl.SCANCODE_DOWN {
		nes.controller.pressButton(controllerButtonDown)
	} else if scancode == sdl.SCANCODE_LEFT {
		nes.controller.pressButton(controllerButtonLeft)
	} else if scancode == sdl.SCANCODE_RIGHT {
		nes.controller.pressButton(controllerButtonRight)	
	}
}

func (nes *NES) controllerButtonReleased(scancode sdl.Scancode) {
	if scancode == sdl.SCANCODE_A {
		nes.controller.releaseButton(controllerButtonA)
	} else if scancode == sdl.SCANCODE_B {
		nes.controller.releaseButton(controllerButtonB)
	} else if scancode == sdl.SCANCODE_Z {
		nes.controller.releaseButton(controllerButtonSelect)					
	} else if scancode == sdl.SCANCODE_X {
		nes.controller.releaseButton(controllerButtonStart)
	} else if scancode == sdl.SCANCODE_UP {
		nes.controller.releaseButton(controllerButtonUp)
	} else if scancode == sdl.SCANCODE_DOWN {
		nes.controller.releaseButton(controllerButtonDown)
	} else if scancode == sdl.SCANCODE_LEFT {
		nes.controller.releaseButton(controllerButtonLeft)
	} else if scancode == sdl.SCANCODE_RIGHT {
		nes.controller.releaseButton(controllerButtonRight)	
	}
}

func drawFrame() {
	texture.Update(nil, renderBuffer[:], windowWidth * argbBytes)
	renderer.Clear()
	renderer.Copy(texture, nil, nil)
	renderer.Present()
}

func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
