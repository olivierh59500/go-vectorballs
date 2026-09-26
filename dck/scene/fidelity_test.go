package scene

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

// This reference follows the former production controller independently of
// geometry.PointSequence. It stays in tests so stage and pose fidelity can be
// checked without importing a graphics device.
type legacyPointAnimation interface{ step(*legacyPointScene) }

type legacyPointScene struct {
	bank      *ShapeManager
	actions   []Action
	shape     *Shape
	position  Vector3
	rotation  Vector3
	rotSpeed  Vector3
	trSpeed   Vector3
	text      int
	anims     []legacyPointAnimation
	index     int
	remaining int
}

func newLegacyPointScene(bank *ShapeManager, actions []Action) *legacyPointScene {
	s := &legacyPointScene{bank: bank, actions: actions, position: Vector3{Z: 850}, index: -1}
	s.next()
	return s
}

func (s *legacyPointScene) next() {
	s.index++
	if s.index >= len(s.actions) {
		s.index = 0
	}
	action := s.actions[s.index]
	s.remaining = int(float64(action.Frames) * FrameRateConversion)
	if action.HasShape {
		s.shape = s.bank.GetCopy(action.Shape)
	}
	if action.HasPos {
		s.position = *action.Pos
	}
	if action.HasInitRot {
		s.rotation = *action.InitRot
	}
	if action.HasRot {
		s.rotSpeed = *action.Rot
	}
	if action.HasTr {
		s.trSpeed = *action.Tr
	}
	if action.HasText {
		s.text = action.TextIndex
	}
	if action.HasAnimTypes {
		s.anims = s.anims[:0]
		for _, name := range action.AnimTypes {
			s.add(name)
		}
	}
	for i := 0; i < action.DropLastAnimation && len(s.anims) > 0; i++ {
		s.anims = s.anims[:len(s.anims)-1]
	}
	for _, name := range action.AppendAnimTypes {
		s.add(name)
	}
}

func (s *legacyPointScene) add(name string) {
	switch name {
	case "Sinus2D":
		s.anims = append(s.anims, &legacyGrid{})
	case "YRotate":
		s.anims = append(s.anims, &legacyOrbit{increment: .025})
	case "Rotors":
		s.anims = append(s.anims, &legacyRotors{})
	case "Bounce":
		s.anims = append(s.anims, newLegacyBounce())
	case "MorphingSphere":
		s.anims = append(s.anims, newLegacyMorph(s.shape.Points, s.bank.shapes["sphere"].Points, int(45*FrameRateConversion)))
	case "MorphingTube":
		s.anims = append(s.anims, newLegacyMorph(s.shape.Points, s.bank.shapes["tube"].Points, int(45*FrameRateConversion)))
	case "MorphingSquare":
		s.anims = append(s.anims, newLegacyMorph(s.shape.Points, s.bank.shapes["square"].Points, int(60*FrameRateConversion)))
	case "MorphingSpaceCube":
		s.anims = append(s.anims, newLegacyMorph(s.shape.Points, s.bank.shapes["spacecube"].Points, int(45*FrameRateConversion)))
	default:
		panic("unknown legacy animation " + name)
	}
}

func (s *legacyPointScene) step() {
	s.remaining--
	if s.remaining <= 0 {
		s.next()
		return
	}
	s.rotation.X += s.rotSpeed.X
	s.rotation.Y += s.rotSpeed.Y
	s.rotation.Z += s.rotSpeed.Z
	s.position.X += s.trSpeed.X
	s.position.Y += s.trSpeed.Y
	s.position.Z += s.trSpeed.Z
	if s.shape != nil {
		for _, animation := range s.anims {
			animation.step(s)
		}
	}
}

type legacyGrid struct{ travel, envelope float64 }

var legacyPhaseSin, legacyPhaseCos = func() ([15]float64, [15]float64) {
	var sin, cos [15]float64
	for i := range sin {
		sin[i], cos[i] = math.Sincos(float64(i) * math.Pi / 10)
	}
	return sin, cos
}()

func (g *legacyGrid) step(s *legacyPointScene) {
	amplitude := 200 * math.Sin(g.envelope)
	sinTravel, cosTravel := math.Sincos(g.travel)
	index := 0
	for row := 0; row < 8; row++ {
		for column := 0; column < 8; column++ {
			if index < len(s.shape.Points) {
				phase := row + column
				s.shape.Points[index].Z = amplitude * (cosTravel*legacyPhaseCos[phase] - sinTravel*legacyPhaseSin[phase])
			}
			index++
		}
	}
	g.travel += math.Pi / 55
	g.envelope += math.Pi / 60
}

type legacyMorph struct {
	steps        []Vector3
	tick, frames int
}

func newLegacyMorph(from, to []Vector3, frames int) *legacyMorph {
	m := &legacyMorph{steps: make([]Vector3, len(from)), frames: frames}
	for i := range from {
		if i < len(to) {
			m.steps[i] = Vector3{X: (to[i].X - from[i].X) / float64(frames),
				Y: (to[i].Y - from[i].Y) / float64(frames), Z: (to[i].Z - from[i].Z) / float64(frames)}
		}
	}
	return m
}

func (m *legacyMorph) step(s *legacyPointScene) {
	if m.tick >= m.frames {
		return
	}
	for i := range s.shape.Points {
		if i < len(m.steps) {
			s.shape.Points[i].X += m.steps[i].X
			s.shape.Points[i].Y += m.steps[i].Y
			s.shape.Points[i].Z += m.steps[i].Z
		}
	}
	m.tick++
}

type legacyOrbit struct{ phase, ratio, increment float64 }

func (o *legacyOrbit) step(s *legacyPointScene) {
	o.phase += math.Pi / 144
	if o.increment > 0 {
		o.ratio += o.increment
		if o.ratio >= 1 {
			o.increment = 0
		}
	} else if o.increment < 0 {
		o.ratio += o.increment
		if o.ratio <= 0 {
			o.increment = 0
		}
	}
	sinPhase, cosPhase := math.Sincos(o.phase)
	s.position = Vector3{X: o.ratio * 100 * cosPhase, Z: 850 + o.ratio*100*sinPhase}
}

type legacyRotors struct{ phase float64 }

func (r *legacyRotors) step(s *legacyPointScene) {
	if len(s.shape.Points) < 8 {
		return
	}
	sinPhase, cosPhase := math.Sincos(r.phase)
	for i, radius := range [...]float64{80, 150, 210, 160} {
		x, z := radius*sinPhase, radius*cosPhase
		s.shape.Points[i*2].X, s.shape.Points[i*2].Z = x, z-100
		s.shape.Points[i*2+1].X, s.shape.Points[i*2+1].Z = -x, -z-100
	}
	r.phase += .08
}

type legacyBounce struct {
	samples []float64
	index   int
}

func newLegacyBounce() *legacyBounce {
	b := &legacyBounce{samples: make([]float64, 0, 540)}
	radius, offset := 1200.0, 210.0
	appendSample := func(angle float64) {
		b.samples = append(b.samples, -(offset-radius*math.Cos(angle*math.Pi/180))/3)
		radius -= 5.7143 / FrameRateConversion
		offset -= 1 / FrameRateConversion
	}
	for cycle := 1; cycle < 4; cycle++ {
		for angle := 0.0; angle < 90; angle += 3 / FrameRateConversion {
			appendSample(angle)
		}
		for angle := 89.0; angle >= 0; angle -= 3 / FrameRateConversion {
			appendSample(angle)
		}
	}
	for angle := 0.0; angle < 90; angle++ {
		appendSample(angle)
	}
	return b
}

func (b *legacyBounce) step(s *legacyPointScene) {
	s.position.Y = b.samples[b.index]
	b.index++
	if b.index == len(b.samples) {
		b.index = 0
	}
}

func TestPointSequenceMatchesFormerControllerOnEveryTick(t *testing.T) {
	actions := AuthoredActions()
	bank := NewShapeManager()
	compiled, err := CompileActions(actions, sourceAnimation)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := geometry.NewPointSequence(geometry.PointSequenceConfig{
		Shapes: bank, Actions: compiled, DurationScale: FrameRateConversion,
		Loop: true, InitialPosition: geometry.Vec3{Z: 850},
	})
	if err != nil {
		t.Fatal(err)
	}
	legacy := newLegacyPointScene(bank, actions)
	frames := 0
	for _, action := range actions {
		frames += int(float64(action.Frames) * FrameRateConversion)
	}
	for tick := 0; tick < frames*2; tick++ {
		if err := shared.Step(); err != nil {
			t.Fatalf("tick %d: %v", tick, err)
		}
		legacy.step()
		state := shared.State()
		if shared.ActionIndex() != legacy.index || shared.Remaining() != legacy.remaining ||
			shared.AnimationCount() != len(legacy.anims) || state.TextIndex != legacy.text {
			t.Fatalf("tick %d stage/cue mismatch: DCK %d/%d/%d/%d, old %d/%d/%d/%d", tick,
				shared.ActionIndex(), shared.Remaining(), shared.AnimationCount(), state.TextIndex,
				legacy.index, legacy.remaining, len(legacy.anims), legacy.text)
		}
		if state.Position != (geometry.Vec3{X: legacy.position.X, Y: legacy.position.Y, Z: legacy.position.Z}) ||
			state.Rotation != (geometry.Vec3{X: legacy.rotation.X, Y: legacy.rotation.Y, Z: legacy.rotation.Z}) {
			t.Fatalf("tick %d model pose mismatch: DCK %+v/%+v, old %+v/%+v", tick,
				state.Position, state.Rotation, legacy.position, legacy.rotation)
		}
		shape, ok := state.Points.(*Shape)
		if !ok || len(shape.Points) != len(legacy.shape.Points) {
			t.Fatalf("tick %d shape point count mismatch", tick)
		}
		for i, point := range shape.Points {
			if point != legacy.shape.Points[i] {
				t.Fatalf("tick %d point %d differs: DCK %+v, old %+v", tick, i, point, legacy.shape.Points[i])
			}
		}
	}
}

func BenchmarkSharedPointSequence(b *testing.B) {
	bank := NewShapeManager()
	actions, err := CompileActions(AuthoredActions(), sourceAnimation)
	if err != nil {
		b.Fatal(err)
	}
	sequence, err := geometry.NewPointSequence(geometry.PointSequenceConfig{
		Shapes: bank, Actions: actions, DurationScale: FrameRateConversion,
		Loop: true, InitialPosition: geometry.Vec3{Z: 850},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := sequence.Step(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormerPointSequence(b *testing.B) {
	legacy := newLegacyPointScene(NewShapeManager(), AuthoredActions())
	b.ReportAllocs()
	for b.Loop() {
		legacy.step()
	}
}
