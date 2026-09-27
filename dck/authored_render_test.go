//go:build dck_authored_rendercheck

package vectorballs

import (
	"fmt"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
)

// This opt-in suite records the authored point sequence without an audio
// device. Before/after directories can be compared byte for byte.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	dir := os.Getenv("DCK_AUTHORED_CAPTURES")
	if dir == "" {
		var err error
		dir, err = os.MkdirTemp("", "dck-vectorballs-authored-")
		if err != nil {
			panic(err)
		}
	}
	frames := []int{0, 1, 30, 60, 120, 239, 240, 600, 1200, 2400, 4800}
	err := capture.Run(capture.Config{Directory: dir, Frames: frames, Width: 640, Height: 480}, func() (ebiten.Game, error) {
		game := &Game{shapeManager: NewShapeManager(), zoomFactor: .35,
			fov: 1450, centerX: 320, centerY: 193, position: Vector3{Z: 850}, dirty: true}
		game.loadImages()
		game.playgroundCanvas = ebiten.NewImage(640, 386)
		game.initReflection()
		game.whiteImage = ebiten.NewImage(1, 1)
		game.whiteImage.Fill(color.White)
		game.initActions()
		if err := game.bindSequence(); err != nil {
			return nil, err
		}
		return game, nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Vectorballs authored captures:", dir)
}
