package vectorballs

import (
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/presets"
)

// These adapters retain the production's action interface while DCK owns the
// point deformations and position clocks. One adapter is reused per animation.
type Sinus2D struct {
	program *geometry.SinusGrid
	points  vectorPointSet
}

func NewSinus2D() *Sinus2D {
	program, err := geometry.NewSinusGrid(presets.VectorballsSinusGrid())
	if err != nil {
		panic(err)
	}
	return &Sinus2D{program: program}
}

func (s *Sinus2D) Run(points []Vector3, _ int) {
	s.points.points = points
	s.program.Step(&s.points)
}

type Rotors struct {
	program *geometry.Rotors
	points  vectorPointSet
}

func NewRotors() *Rotors {
	program, err := geometry.NewRotors(presets.VectorballsRotors())
	if err != nil {
		panic(err)
	}
	return &Rotors{program: program}
}

func (r *Rotors) Run(points []Vector3, _ int) {
	r.points.points = points
	r.program.Step(&r.points)
}

type YRotate struct {
	program  *geometry.YOrbit
	position Vector3
}

func NewYRotate() *YRotate {
	program, err := geometry.NewYOrbit(presets.VectorballsYOrbit())
	if err != nil {
		panic(err)
	}
	return &YRotate{program: program}
}

func (y *YRotate) Run(_ []Vector3, _ int) {
	position := y.program.Step()
	y.position = Vector3{X: position.X, Y: position.Y, Z: position.Z}
}

func (y *YRotate) GetPosition() Vector3 { return y.position }

type Bounce struct{ program *geometry.BounceCurve }

func NewBounce() *Bounce {
	program, err := geometry.NewBounceCurve(presets.VectorballsBounce())
	if err != nil {
		panic(err)
	}
	return &Bounce{program: program}
}

func (b *Bounce) Run(_ []Vector3, _ int) { b.program.Step() }
func (b *Bounce) GetBounceY() float64    { return b.program.At() }
