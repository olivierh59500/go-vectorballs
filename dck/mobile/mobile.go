// Package vectorballsmobile exposes the game to Ebitengine's Android view.
package vectorballsmobile

import (
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/sprites"
	vectorballs "go-vectorballs/dck"
)

var preview atomic.Value

// ConfigurePreview chooses a DCK object before Android creates its first game
// frame. An empty name leaves the authored choreography in place. The returned
// string is empty on success so the Java activity can report invalid extras.
func ConfigurePreview(name, fill string, segments int, size float64, image int) string {
	preview.Store(vectorballs.ObjectOptions{})
	if name == "" {
		return ""
	}
	var material sprites.Fill
	switch fill {
	case "", "edges":
		material = sprites.Edges
	case "surface":
		material = sprites.Surface
	case "solid":
		material = sprites.Solid
	default:
		return "unknown object fill"
	}
	options := vectorballs.ObjectOptions{Name: name, Fill: material, Segments: segments, Size: size, Image: image}
	config, err := presets.VectorballsProjectedObject(options.Name, options.Fill, options.Segments, options.Size, options.Image)
	if err != nil {
		return err.Error()
	}
	if _, err := sprites.NewProjectedObject(config); err != nil {
		return err.Error()
	}
	preview.Store(options)
	return ""
}

func init() {
	ebiten.SetScreenClearedEveryFrame(false)
	mobile.SetGame(&game{})
}

// game delays resource creation until Ebitengine has an Android context and a
// rendering surface. Package initializers run while go.Seq is still loading,
// before MainActivity can call Seq.setContext.
type game struct {
	delegate *vectorballs.Game
}

func (g *game) Update() error {
	if g.delegate == nil {
		g.delegate = vectorballs.NewGame()
		if value := preview.Load(); value != nil {
			if options := value.(vectorballs.ObjectOptions); options.Name != "" {
				if err := g.delegate.SetObject(options); err != nil {
					return err
				}
			}
		}
	}
	return g.delegate.Update()
}

func (g *game) Draw(screen *ebiten.Image) {
	if g.delegate != nil {
		g.delegate.Draw(screen)
	}
}

func (g *game) Layout(_, _ int) (int, int) {
	return 640, 480
}

// Dummy makes this package bindable by gomobile.
func Dummy() {}
