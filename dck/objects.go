package vectorballs

import (
	"fmt"

	"github.com/olivierh59500/democonstructionkit/motion"
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
		g.objectRest = nil
		g.objectFlag = nil
		if err := g.sequence.Reset(); err != nil {
			return err
		}
		g.syncSequence()
		return nil
	}
	var points []sprites.Point
	var err error
	var flag *sprites.Flag
	switch c.Name {
	case "cube":
		points, err = sprites.Cube(sprites.CubeConfig{Size: c.Size, Segments: c.Segments, Fill: c.Fill, Image: c.Image})
	case "pyramid":
		points, err = sprites.Pyramid(sprites.PyramidConfig{Width: c.Size, Height: c.Size, Segments: c.Segments, Fill: c.Fill, Image: c.Image})
	case "plane", "flag":
		points, err = sprites.Plane(sprites.PlaneConfig{Width: c.Size, Height: c.Size * .65, Columns: c.Segments, Rows: c.Segments, Image: c.Image})
		if c.Name == "flag" {
			flag = &sprites.Flag{Width: c.Size, PinLeft: true, Wave: motion.Wave{Amplitude: c.Size * .15, Spatial: 7 / c.Size, Speed: 3}, RowPhase: 2 / c.Size}
		}
	default:
		return fmt.Errorf("vectorballs: unknown object %q", c.Name)
	}
	if err != nil {
		return err
	}
	if c.Image < 0 || c.Image > 120 {
		return fmt.Errorf("vectorballs: ball index must be within [0,120]")
	}
	g.objectRest = points
	g.objectAnimated = make([]sprites.Point, len(points))
	g.objectFlag = flag
	g.currentShape = &Shape{Points: make([]Vector3, len(points))}
	g.rotation = Vector3{}
	g.position = Vector3{Z: 850}
	g.zoomFactor = .55
	g.currentText = -1
	g.updateObject()
	return nil
}

func (g *Game) updateObject() {
	copy(g.objectAnimated, g.objectRest)
	if g.objectFlag != nil {
		g.objectFlag.Apply(g.objectAnimated, g.objectRest, float64(g.frameCount)/60)
	}
	for i, p := range g.objectAnimated {
		g.currentShape.Points[i] = Vector3{X: p.X, Y: p.Y, Z: p.Z, Img: p.Image}
	}
	g.rotation.X += .012
	g.rotation.Y += .017
	g.rotation.Z += .005
	g.dirty = true
}
