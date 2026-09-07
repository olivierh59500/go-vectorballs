package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	vectorballs "go-vectorballs"
)

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Red Sector Vectorballs Demo by TLB (ported in golang by bilizir from DMA)")

	game := vectorballs.NewGame()
	defer game.Cleanup()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
