package vectorballs

import (
	"math"
	"testing"
)

func TestPointAnimationAdaptersKeepAuthoredFirstFrames(t *testing.T) {
	points := make([]Vector3, 64)
	for i := range points {
		points[i] = Vector3{X: float64(i), Y: float64(i * 7), Z: 12, Img: 37}
	}
	sinus := NewSinus2D()
	sinus.Run(points, 0)
	if points[0].Z != 0 || points[63].Z != 0 || points[0].Img != 37 {
		t.Fatalf("first sinus grid frame changed artwork or height: %+v", points[0])
	}
	sinus.Run(points, 1)
	wantZ := 200 * math.Sin(math.Pi/60) * math.Cos(math.Pi/55)
	if points[0].Z != wantZ {
		t.Fatalf("second sinus grid Z = %v, want %v", points[0].Z, wantZ)
	}

	rotors := NewRotors()
	rotors.Run(points, 0)
	if points[0].X != 0 || points[0].Z != -20 || points[1].Z != -180 ||
		points[0].Y != 0 || points[1].Y != 7 || points[0].Img != 37 {
		t.Fatalf("first rotor pose differs: %+v %+v", points[0], points[1])
	}

	orbit := NewYRotate()
	orbit.Run(points, 0)
	wantOrbit := Vector3{X: .025 * 100 * math.Cos(math.Pi/144), Z: 850 + .025*100*math.Sin(math.Pi/144)}
	if got := orbit.GetPosition(); got != wantOrbit {
		t.Fatalf("first Y orbit %+v, want %+v", got, wantOrbit)
	}

	bounce := NewBounce()
	bounce.Run(points, 0)
	if got := bounce.GetBounceY(); got != 330 || points[0].Img != 37 {
		t.Fatalf("first bounce = %v, ball image %d", got, points[0].Img)
	}
}
