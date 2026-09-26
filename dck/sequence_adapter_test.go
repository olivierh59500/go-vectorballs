package vectorballs

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestAuthoredPointSequenceKeepsEveryStageBoundary(t *testing.T) {
	game := &Game{shapeManager: NewShapeManager()}
	game.initActions()
	if err := game.bindSequence(); err != nil {
		t.Fatal(err)
	}
	sequence := game.sequence
	if len(game.actions) < 30 || sequence.ActionIndex() != 0 || sequence.Remaining() != int(float64(game.actions[0].Frames)*frameRateConversion) {
		t.Fatalf("initial stage = %d, remaining %d", sequence.ActionIndex(), sequence.Remaining())
	}
	first := sequence.State()
	firstXYZ := first.Points.XYZ(0)
	if first.Points.Len() != 64 || first.Position != (geometry.Vec3{Y: 320, Z: 850}) ||
		first.Rotation != (geometry.Vec3{X: 30 * (math.Pi / 180), Y: 30 * (math.Pi / 180), Z: 30 * (math.Pi / 180)}) ||
		first.TextIndex != 0 || sequence.AnimationCount() != 0 {
		t.Fatalf("first authored stage = %+v", first)
	}

	count := sequence.AnimationCount()
	for index, action := range game.actions {
		if sequence.ActionIndex() != index {
			t.Fatalf("stage %d entered action %d", index, sequence.ActionIndex())
		}
		wantFrames := int(float64(action.Frames) * frameRateConversion)
		if sequence.Remaining() != wantFrames {
			t.Fatalf("stage %d duration = %d, want %d", index, sequence.Remaining(), wantFrames)
		}
		if action.HasAnimTypes {
			count = len(action.AnimTypes)
		}
		count = max(0, count-action.DropLastAnimation) + len(action.AppendAnimTypes)
		if sequence.AnimationCount() != count {
			t.Fatalf("stage %d active animations = %d, want %d", index, sequence.AnimationCount(), count)
		}
		if action.HasShape {
			shape := game.shapeManager.GetCopy(action.Shape)
			if sequence.State().Points.Len() != len(shape.Points) {
				t.Fatalf("stage %d shape point count changed", index)
			}
		}
		if action.HasText && sequence.State().TextIndex != action.TextIndex {
			t.Fatalf("stage %d caption cue = %d, want %d", index, sequence.State().TextIndex, action.TextIndex)
		}
		for tick := 0; tick < wantFrames; tick++ {
			if err := sequence.Step(); err != nil {
				t.Fatalf("stage %d tick %d: %v", index, tick, err)
			}
		}
	}
	if sequence.ActionIndex() != 0 || sequence.Finished() {
		t.Fatal("authored action program did not loop to its first stage")
	}
	if err := sequence.Reset(); err != nil {
		t.Fatal(err)
	}
	if sequence.ActionIndex() != 0 || sequence.State().Points.XYZ(0) != firstXYZ {
		t.Fatal("restarting the script retained a mutated shape")
	}
}
