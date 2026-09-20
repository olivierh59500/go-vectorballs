//go:build dck_rendercheck

package vectorballs

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

var objectCases = []ObjectOptions{
	{Name: "cube", Fill: sprites.Edges, Segments: 4, Image: 120, Size: 640},
	{Name: "cube", Fill: sprites.Surface, Segments: 4, Image: 120, Size: 640},
	{Name: "cube", Fill: sprites.Solid, Segments: 4, Image: 120, Size: 640},
	{Name: "pyramid", Fill: sprites.Edges, Segments: 6, Image: 120, Size: 640},
	{Name: "pyramid", Fill: sprites.Solid, Segments: 6, Image: 120, Size: 640},
	{Name: "plane", Segments: 8, Image: 120, Size: 640},
	{Name: "flag", Segments: 12, Image: 120, Size: 640},
}

type objectCheck struct {
	*Game
	tick int
	err  error
}

func (c *objectCheck) Update() error {
	if c.err != nil {
		return c.err
	}
	c.tick++
	if c.tick%60 == 0 && c.tick/60 < len(objectCases) {
		if err := c.SetObject(objectCases[c.tick/60]); err != nil {
			return err
		}
	}
	return c.Game.Update()
}
func (c *objectCheck) Draw(dst *ebiten.Image) {
	c.Game.Draw(dst)
	if c.tick%60 != 30 {
		return
	}
	pixels := make([]byte, 640*386*4)
	c.playgroundCanvas.ReadPixels(pixels)
	visible := 0
	for i := 3; i < len(pixels); i += 4 {
		if pixels[i] != 0 {
			visible++
		}
	}
	if visible < 100 {
		c.err = fmt.Errorf("object %d has only %d visible pixels", c.tick/60, visible)
	}
}

func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	dir := os.Getenv("DCK_OBJECT_CAPTURES")
	if dir == "" {
		var err error
		dir, err = os.MkdirTemp("", "dck-objects-")
		if err != nil {
			panic(err)
		}
	}
	var check *objectCheck
	err := capture.Run(capture.Config{Directory: dir, Frames: []int{30, 90, 150, 210, 270, 330, 390, 420}, Width: 640, Height: 480}, func() (ebiten.Game, error) {
		g := &Game{shapeManager: NewShapeManager(), zoomFactor: .35, fov: 1450, centerX: 320, centerY: 193, position: Vector3{Z: 850}, dirty: true}
		g.loadImages()
		g.playgroundCanvas = ebiten.NewImage(640, 386)
		g.reflectionSource = g.playgroundCanvas.SubImage(image.Rect(0, 288, 640, 368)).(*ebiten.Image)
		g.whiteImage = ebiten.NewImage(1, 1)
		g.whiteImage.Fill(color.White)
		if err := g.SetObject(objectCases[0]); err != nil {
			return nil, err
		}
		check = &objectCheck{Game: g}
		return check, nil
	})
	if err == nil && check != nil {
		err = check.err
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Vectorball objects verified; captures:", dir)
}
