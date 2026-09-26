package vectorballs

import (
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/presets"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

type ObjectOptions struct {
	Name            string
	Fill            sprites.Fill
	Segments, Image int
	Size            float64
}

// SetObject replaces the action script with a rotating construction-kit object.
// An empty name restores the original choreography. The ball atlas, projection,
// stage and reflection remain the production's own presentation.
func (g *Game) SetObject(c ObjectOptions) error {
	if c.Name == "" {
		g.object = nil
		if err := g.sequence.Reset(); err != nil {
			return err
		}
		g.syncSequence()
		return nil
	}
	config, err := presets.VectorballsProjectedObject(c.Name, c.Fill, c.Segments, c.Size, c.Image)
	if err != nil {
		return err
	}
	object, err := sprites.NewProjectedObject(config)
	if err != nil {
		return err
	}
	if err := object.Update(kit.Frame{Time: float64(g.frameCount) / 60}); err != nil {
		return err
	}
	g.object = object
	g.currentText = -1
	g.dirty = true
	return nil
}
