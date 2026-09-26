package scene

import (
	"fmt"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

func sourceAnimation(name string) (geometry.PointAnimationSpec, error) {
	switch name {
	case "Sinus2D":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimSinusGrid,
			SinusGrid: geometry.SinusGridConfig{Columns: 8, Rows: 8, Amplitude: 200,
				SpatialNumerator: math.Pi, SpatialDivisor: 10,
				TravelStep: math.Pi / 55, EnvelopeStep: math.Pi / 60}}, nil
	case "YRotate":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimYOrbit,
			YOrbit: geometry.YOrbitConfig{Center: geometry.Vec3{Z: 850}, Radius: 100,
				PhaseStep: math.Pi / 144, RatioStep: .025, Min: 0, Max: 1}}, nil
	case "Rotors":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimRotors,
			Rotors: geometry.RotorConfig{Radii: []float64{80, 150, 210, 160}, Offset: 100,
				PhaseStep: .08, RequireFull: true}}, nil
	case "Bounce":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimBounce,
			Bounce: geometry.BounceCurveConfig{Radius: 1200, Offset: 210,
				RadiusDecay: 5.7143 / FrameRateConversion, OffsetDecay: 1 / FrameRateConversion,
				StepDegrees: 3 / FrameRateConversion, BackwardStart: 89,
				FinalStepDegrees: 1, Loops: 3, Divisor: 3}}, nil
	case "MorphingSphere":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "sphere"}, nil
	case "MorphingTube":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "tube"}, nil
	case "MorphingSquare":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "square"}, nil
	case "MorphingSpaceCube":
		return geometry.PointAnimationSpec{Kind: geometry.PointAnimMorph, Target: "spacecube"}, nil
	default:
		return geometry.PointAnimationSpec{}, fmt.Errorf("unknown animation %q", name)
	}
}

func TestAuthoredShapesRemainIndependentAndIndexed(t *testing.T) {
	shapes := NewShapeManager()
	for name, want := range map[string]int{"square": 64, "tube": 64, "sphere": 64, "heli": 62, "cube": 27, "spacecube": 27} {
		shape, err := shapes.Clone(name)
		if err != nil || shape.Len() != want {
			t.Fatalf("shape %q has %d points, want %d: %v", name, shape.Len(), want, err)
		}
	}
	first := shapes.GetCopy("square")
	original := first.Points[0]
	first.Points[0].X = 999
	if second := shapes.GetCopy("square"); second.Points[0] != original || original.Img != 29 {
		t.Fatal("shape clone changed the authored point or ball index")
	}
}

func TestCompileActionsDistinguishesAbsentAndExplicitFields(t *testing.T) {
	zero := &Vector3{}
	actions, err := CompileActions([]Action{
		{Frames: 2, Pos: zero, HasPos: true, AnimTypes: []string{}, HasAnimTypes: true},
		{Frames: 3, Pos: zero},
		{Frames: 4, AppendAnimTypes: []string{"YRotate"}, DropLastAnimation: 1},
	}, sourceAnimation)
	if err != nil {
		t.Fatal(err)
	}
	if actions[0].Position == nil || *actions[0].Position != (geometry.Vec3{}) || !actions[0].SetAnimations ||
		actions[1].Position != nil || actions[1].SetAnimations || actions[2].DropLast != 1 ||
		len(actions[2].AppendAnimations) != 1 {
		t.Fatalf("action field presence changed: %+v", actions)
	}
}

func TestFullAuthoredSequenceRunsThroughEveryStageAndLoops(t *testing.T) {
	shapes := NewShapeManager()
	authored := AuthoredActions()
	if len(authored) < 30 {
		t.Fatalf("only %d authored actions", len(authored))
	}
	actions, err := CompileActions(authored, sourceAnimation)
	if err != nil {
		t.Fatal(err)
	}
	sequence, err := geometry.NewPointSequence(geometry.PointSequenceConfig{
		Shapes: shapes, Actions: actions, DurationScale: FrameRateConversion,
		Loop: true, InitialPosition: geometry.Vec3{Z: 850},
	})
	if err != nil {
		t.Fatal(err)
	}
	firstPoint := shapes.GetCopy("square").Points[0]
	count := 0
	for index, action := range authored {
		if sequence.ActionIndex() != index || sequence.Remaining() != int(float64(action.Frames)*FrameRateConversion) {
			t.Fatalf("stage %d entered as %d with %d ticks", index, sequence.ActionIndex(), sequence.Remaining())
		}
		if action.HasAnimTypes {
			count = len(action.AnimTypes)
		}
		count = max(0, count-action.DropLastAnimation) + len(action.AppendAnimTypes)
		if sequence.AnimationCount() != count {
			t.Fatalf("stage %d has %d animations, want %d", index, sequence.AnimationCount(), count)
		}
		if action.HasShape {
			template := shapes.GetCopy(action.Shape)
			if sequence.State().Points.Len() != template.Len() || sequence.State().Points.XYZ(0) != template.XYZ(0) {
				t.Fatalf("stage %d did not clone the authored %q shape", index, action.Shape)
			}
		}
		if action.HasText && sequence.State().TextIndex != action.TextIndex {
			t.Fatalf("stage %d caption cue changed", index)
		}
		for tick := 0; tick < int(float64(action.Frames)*FrameRateConversion); tick++ {
			if err := sequence.Step(); err != nil {
				t.Fatalf("stage %d tick %d: %v", index, tick, err)
			}
		}
	}
	if sequence.ActionIndex() != 0 || sequence.Finished() || shapes.GetCopy("square").Points[0] != firstPoint {
		t.Fatal("full authored loop changed its immutable shape template")
	}
}
