package vectorballs

import "testing"

func TestMorphingToKeepsBallImagesAndContinuousHandoff(t *testing.T) {
	points := []Vector3{{X: 1.3, Y: -2, Z: 3, Img: 37}, {X: 8, Img: 120}}
	target := []Vector3{{X: 10.2, Y: 9, Z: -4, Img: 1}}
	morph := NewMorphingTo(points, target, 7)
	want := append([]Vector3(nil), points...)
	for tick := 0; tick < 7; tick++ {
		morph.Run(points, tick)
		want[0].X += (target[0].X - 1.3) / 7
		want[0].Y += (target[0].Y + 2) / 7
		want[0].Z += (target[0].Z - 3) / 7
		for i := range points {
			if points[i] != want[i] {
				t.Fatalf("tick %d ball %d = %+v, want %+v", tick, i, points[i], want[i])
			}
		}
	}
	before := append([]Vector3(nil), points...)
	second := NewMorphingTo(points, []Vector3{{X: -10, Img: 5}}, 5)
	if points[0] != before[0] || points[1] != before[1] {
		t.Fatal("selecting the next morph changed the current ball positions")
	}
	second.Run(points, 0)
	if points[0].X != before[0].X+(-10-before[0].X)/5 || points[0].Img != 37 || points[1] != before[1] {
		t.Fatalf("morph handoff jumped or replaced ball images: %+v", points)
	}
}
