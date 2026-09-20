package main

import (
	"flag"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/sprites"
	vectorballs "go-vectorballs/dck"
)

func main() {
	object := flag.String("object", "", "optional object: cube, pyramid, plane or flag")
	fill := flag.String("fill", "edges", "cube/pyramid filling: edges, surface or solid")
	segments := flag.Int("segments", 6, "edge subdivisions")
	size := flag.Float64("size", 640, "object size in model units")
	ball := flag.Int("ball", 120, "ball sprite index")
	flag.Parse()
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Red Sector Vectorballs Demo by TLB (ported in golang by bilizir from DMA)")
	ebiten.SetScreenClearedEveryFrame(false)

	game := vectorballs.NewGame()
	defer game.Cleanup()
	if *object != "" {
		fills := map[string]sprites.Fill{"edges": sprites.Edges, "surface": sprites.Surface, "solid": sprites.Solid}
		mode, ok := fills[*fill]
		if !ok {
			log.Fatalf("unknown fill %q", *fill)
		}
		if err := game.SetObject(vectorballs.ObjectOptions{Name: *object, Fill: mode, Segments: *segments, Size: *size, Image: *ball}); err != nil {
			log.Fatal(err)
		}
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
