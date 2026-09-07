// Package vectorballsmobile exposes the game to Ebitengine's Android view.
package vectorballsmobile

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/mobile"
	vectorballs "go-vectorballs"
)

func init() {
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
