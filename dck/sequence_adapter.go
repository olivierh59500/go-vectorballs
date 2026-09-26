package vectorballs

import (
	"fmt"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/presets"
	"go-vectorballs/dck/scene"
)

func authoredAnimation(name string) (geometry.PointAnimationSpec, error) {
	switch name {
	case "Sinus2D":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimSinusGrid, SinusGrid: presets.VectorballsSinusGrid()}, nil
	case "YRotate":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimYOrbit, YOrbit: presets.VectorballsYOrbit()}, nil
	case "Rotors":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimRotors, Rotors: presets.VectorballsRotors()}, nil
	case "Bounce":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimBounce, Bounce: presets.VectorballsBounce()}, nil
	case "MorphingSphere":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "sphere"}, nil
	case "MorphingTube":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "tube"}, nil
	case "MorphingSquare":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "square"}, nil
	case "MorphingSpaceCube":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "spacecube"}, nil
	default:
		return geometry.PointAnimationSpec{}, fmt.Errorf("vectorballs: unknown animation %q", name)
	}
}

func (g *Game) bindSequence() error {
	actions, err := scene.CompileActions(g.actions, authoredAnimation)
	if err != nil {
		return err
	}
	sequence, err := geometry.NewPointSequence(geometry.PointSequenceConfig{
		Shapes: g.shapeManager, Actions: actions, DurationScale: frameRateConversion,
		Loop: true, InitialPosition: geometry.Vec3{Z: 850},
	})
	if err != nil {
		return err
	}
	g.sequence = sequence
	g.syncSequence()
	return nil
}

func (g *Game) syncSequence() {
	state := g.sequence.State()
	g.currentShape, _ = state.Points.(*Shape)
	g.position = Vector3{X: state.Position.X, Y: state.Position.Y, Z: state.Position.Z}
	g.rotation = Vector3{X: state.Rotation.X, Y: state.Rotation.Y, Z: state.Rotation.Z}
	g.currentText = state.TextIndex
}
