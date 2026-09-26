package vectorballs

import "github.com/olivierh59500/democonstructionkit/geometry"

// vectorPointSet adapts the authored ball records without copying their image
// indices or allocating a new point slice on every morph frame.
type vectorPointSet struct{ points []Vector3 }

func (s *vectorPointSet) Len() int { return len(s.points) }
func (s *vectorPointSet) XYZ(index int) geometry.Vec3 {
	p := s.points[index]
	return geometry.Vec3{X: p.X, Y: p.Y, Z: p.Z}
}
func (s *vectorPointSet) SetXYZ(index int, value geometry.Vec3) {
	p := &s.points[index]
	p.X, p.Y, p.Z = value.X, value.Y, value.Z
}

// MorphingTo keeps the production's animation interface while DCK owns the
// geometry, frame count and handoff-safe incremental motion.
type MorphingTo struct {
	program *geometry.PointMorph
	points  vectorPointSet
}

func NewMorphingTo(from, to []Vector3, frames int) *MorphingTo {
	program, err := geometry.NewPointMorph(
		&vectorPointSet{points: from}, &vectorPointSet{points: to},
		geometry.PointMorphConfig{Frames: frames, Tail: geometry.MorphHold},
	)
	if err != nil {
		panic(err) // All authored morph durations and points are validated data.
	}
	return &MorphingTo{program: program}
}

func (m *MorphingTo) Run(points []Vector3, _ int) {
	m.points.points = points
	m.program.Step(&m.points)
}
