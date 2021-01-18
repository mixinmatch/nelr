package main

type GameController struct {
	buttonStates byte // A-B-Se-St-U-D-L-R https://wiki.nesdev.com/w/index.php/Controller_reading_code
	strobe bool
}
//https://wiki.nesdev.com/w/index.php/Standard_controller
const(
	controllerButtonA = 0
	controllerButtonB = 1
	controllerButtonSelect = 2
	controllerButtonStart = 3
	controllerButtonUp = 4
	controllerButtonDown = 5
	controllerButtonLeft = 6
	controllerButtonRight = 7
)


func MakeNewGameController() *GameController {
	return &GameController {
		buttonStates: 0,
		strobe: false,
	}
}
func (g *GameController) Write(value byte) {
	g.strobe = value&1 == 1
}

func (g *GameController) Read() byte {
	if g.strobe {
		buttonAState := g.buttonStates&1
		return buttonAState
	} else {
		btnState := (g.buttonStates >> 7) //TODO-check
		g.buttonStates <<= 1
		g.buttonStates |= btnState
		return 0x40 | btnState
	}
}

func (g *GameController) pressButton(button byte) {
	buttonBit := byte(7)-button 
	g.buttonStates |= (1<<buttonBit)  	
}

func (g *GameController) releaseButton(button byte) {
	buttonBit := byte(7)-button 
	mask := byte(0xFF)^(1<<buttonBit)
	g.buttonStates &= mask
}
