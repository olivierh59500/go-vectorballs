package vectorballs

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/olivierh59500/ym-player/pkg/stsound"
)

const (
	screenWidth  = 640
	screenHeight = 480

	// Demo timings (60 fps web vs 25 fps Atari ST)
	frameRateConversion = 2.4

	// Android and most current desktop audio devices use 48 kHz natively.
	audioSampleRate = 48_000
)

// Embedded assets
var (
	//go:embed assets/AllBalls.png
	ballsData []byte
	//go:embed assets/text.png
	textData []byte
	//go:embed assets/Mindbomb.ym
	musicData []byte

	sinusPhaseSin, sinusPhaseCos = func() ([15]float64, [15]float64) {
		var sin, cos [15]float64
		for i := range sin {
			sin[i], cos[i] = math.Sincos(float64(i) * math.Pi / 10)
		}
		return sin, cos
	}()
)

// Vector3 represents a 3D point
type Vector3 struct {
	X, Y, Z float64
	Img     int // Ball image index
}

// Point3D represents a transformed 3D point with depth
type Point3D struct {
	Depth float64
	X2D   float64
	Y2D   float64
	Img   int
}

type matrix3 [9]float64

func newRotationMatrix(rotation Vector3, scale float64) matrix3 {
	sinX, cosX := math.Sincos(rotation.X)
	sinY, cosY := math.Sincos(rotation.Y)
	sinZ, cosZ := math.Sincos(rotation.Z)
	return matrix3{
		cosY * cosZ * scale,
		(sinX*sinY*cosZ - cosX*sinZ) * scale,
		(cosX*sinY*cosZ + sinX*sinZ) * scale,
		cosY * sinZ * scale,
		(sinX*sinY*sinZ + cosX*cosZ) * scale,
		(cosX*sinY*sinZ - sinX*cosZ) * scale,
		-sinY * scale,
		sinX * cosY * scale,
		cosX * cosY * scale,
	}
}

func (m matrix3) apply(point Vector3) (x, y, z float64) {
	return point.X*m[0] + point.Y*m[1] + point.Z*m[2],
		point.X*m[3] + point.Y*m[4] + point.Z*m[5],
		point.X*m[6] + point.Y*m[7] + point.Z*m[8]
}

// Shape represents a 3D shape made of points
type Shape struct {
	Points []Vector3
}

// ShapeManager manages all 3D shapes
type ShapeManager struct {
	shapes map[string]*Shape
}

// NewShapeManager creates and initializes all shapes
func NewShapeManager() *ShapeManager {
	sm := &ShapeManager{
		shapes: make(map[string]*Shape, 12),
	}
	sm.initShapes()
	return sm
}

// parseShape parses shape data string into Vector3 points
func parseShape(data string, defaultBall int) *Shape {
	parts := strings.Split(data, "|")
	points := make([]Vector3, 0, len(parts))

	// Scale factor to space out the balls more
	const scaleFactor = 1.6

	for _, part := range parts {
		coords := strings.Split(part, ",")
		if len(coords) < 3 {
			continue
		}

		x, _ := strconv.ParseFloat(coords[0], 64)
		y, _ := strconv.ParseFloat(coords[1], 64)
		z, _ := strconv.ParseFloat(coords[2], 64)
		img := defaultBall

		if len(coords) >= 4 {
			img, _ = strconv.Atoi(coords[3])
		}

		points = append(points, Vector3{X: x * scaleFactor, Y: y * scaleFactor, Z: z * scaleFactor, Img: img})
	}

	return &Shape{Points: points}
}

// initShapes initializes all shape definitions
func (sm *ShapeManager) initShapes() {
	sm.shapes["square"] = parseShape("-280,280,0,29|-200,280,0,29|-120,280,0,29|-40,280,0,29|40,280,0,29|120,280,0,29|200,280,0,29|280,280,0,29|-280,200,0,29|-200,200,0,37|-120,200,0,37|-40,200,0,37|40,200,0,37|120,200,0,37|200,200,0,37|280,200,0,29|-280,120,0,29|-200,120,0,29|-120,120,0,37|-40,120,0,29|40,120,0,29|120,120,0,37|200,120,0,29|280,120,0,29|-280,40,0,29|-200,40,0,29|-120,40,0,37|-40,40,0,29|40,40,0,29|120,40,0,37|200,40,0,29|280,40,0,29|-280,-40,0,29|-200,-40,0,29|-120,-40,0,37|-40,-40,0,29|40,-40,0,29|120,-40,0,37|200,-40,0,29|280,-40,0,29|-280,-120,0,29|-200,-120,0,29|-120,-120,0,37|-40,-120,0,29|40,-120,0,29|120,-120,0,37|200,-120,0,29|280,-120,0,29|-280,-200,0,29|-200,-200,0,37|-120,-200,0,37|-40,-200,0,37|40,-200,0,37|120,-200,0,37|200,-200,0,37|280,-200,0,29|-280,-280,0,29|-200,-280,0,29|-120,-280,0,29|-40,-280,0,29|40,-280,0,29|120,-280,0,29|200,-280,0,29|280,-280,0,29", 29)

	sm.shapes["tube"] = parseShape("0,280,80,29|56,280,56,29|80,280,0,29|56,280,-57,29|0,280,-80,29|-57,280,-57,29|-80,280,-1,29|-57,280,56,29|0,200,80,29|56,200,56,37|80,200,0,37|56,200,-57,37|0,200,-80,37|-57,200,-57,37|-80,200,-1,37|-57,200,56,29|0,120,80,29|56,120,56,29|80,120,0,37|56,120,-57,29|0,120,-80,29|-57,120,-57,37|-80,120,-1,29|-57,120,56,29|0,40,80,29|56,40,56,29|80,40,0,37|56,40,-57,29|0,40,-80,29|-57,40,-57,37|-80,40,-1,29|-57,40,56,29|0,-40,80,29|56,-40,56,29|80,-40,0,37|56,-40,-57,29|0,-40,-80,29|-57,-40,-57,37|-80,-40,-1,29|-57,-40,56,29|0,-120,80,29|56,-120,56,29|80,-120,0,37|56,-120,-57,29|0,-120,-80,29|-57,-120,-57,37|-80,-120,-1,29|-57,-120,56,29|0,-200,80,29|56,-200,56,37|80,-200,0,37|56,-200,-57,37|0,-200,-80,37|-57,-200,-57,37|-80,-200,-1,37|-57,-200,56,29|0,-280,80,29|56,-280,56,29|80,-280,0,29|56,-280,-57,29|0,-280,-80,29|-57,-280,-57,29|-80,-280,-1,29|-57,-280,56,29", 29)

	sm.shapes["sphere"] = parseShape("0,85,234|60,60,234|85,0,234|60,-61,234|0,-86,234|-61,-61,234|-86,-1,234|-61,60,234|0,160,191|113,113,191|160,0,191|113,-114,191|0,-161,191|-114,-114,191|-161,-1,191|-114,113,191|0,216,125|152,152,125|216,0,125|152,-154,125|0,-217,125|-154,-154,125|-217,-1,125|-154,152,125|0,246,43|173,173,43|246,0,43|173,-175,43|0,-247,43|-175,-175,43|-247,-1,43|-175,173,43|0,246,-44|173,173,-44|246,0,-44|173,-175,-44|0,-247,-44|-175,-175,-44|-247,-1,-44|-175,173,-44|0,216,-126|152,152,-126|216,0,-126|152,-154,-126|0,-217,-126|-154,-154,-126|-217,-1,-126|-154,152,-126|0,160,-192|113,113,-192|160,0,-192|113,-114,-192|0,-161,-192|-114,-114,-192|-161,-1,-192|-114,113,-192|0,85,-235|60,60,-235|85,0,-235|60,-61,-235|0,-86,-235|-61,-61,-235|-86,-1,-235|-61,60,-235", 37)

	sm.shapes["heli"] = parseShape("0,-200,-20,85|0,-200,-180,85|0,-200,50,84|0,-200,-250,84|0,-200,110,83|0,-200,-310,83|0,-200,160,82|0,-200,-360,82|240,100,160|240,100,80|240,100,0|240,100,-70,76|240,100,-130,75|240,100,-180,74|-240,100,160|-240,100,80|-240,100,0|-240,100,-70,76|-240,100,-130,75|-240,100,-180,74|-180,60,40,74|-180,60,-40,74|180,60,40,74|180,60,-40,74|0,10,-180|0,30,-100|0,30,-20|0,30,60|0,10,120,75|80,10,-180,75|80,30,-100|80,30,-20|80,30,60|80,10,120,75|-80,10,-180,75|-80,30,-100|-80,30,-20|-80,30,60|-80,10,120,75|80,-50,-180|80,-50,-100|80,-50,-20|80,-50,60,21|0,-50,60,21|-80,-50,-180|-80,-50,-100|-80,-50,-20|-80,-50,60,21|70,-120,-180|80,-130,-100|80,-130,-20,21|-70,-120,-180|-80,-130,-100|-80,-130,-20,21|0,-130,-180|0,-130,-100,75|0,-130,-20,21|0,-50,-220|0,-60,-290,76|0,-70,-350,75|0,-80,-400,74|0,-170,-100,84", 77)

	sm.shapes["animal"] = parseShape("0,0,-70,94|0,0,70,94|0,0,150,82|0,-40,190,81|0,-70,230,80|0,60,-130,84|0,110,-170,84|0,160,-220,84|0,210,-300,94|50,260,-300,82|-50,260,-300,82|70,280,-300,81|-70,280,-300,81|90,295,-300,80|-90,295,-300,80|30,220,-360,82|-30,220,-360,82|90,-40,70,84|110,-90,90,84|120,-150,110,84|120,-210,125,84|120,-300,140,94|-90,-40,70,84|-110,-90,90,84|-120,-150,110,84|-120,-210,125,84|-120,-300,140,94|90,-40,-70,84|110,-90,-80,84|120,-150,-90,84|120,-210,-95,84|120,-300,-100,94|-90,-40,-70,84|-110,-90,-80,84|-120,-150,-90,84|-120,-210,-95,84|-120,-300,-100,94", 94)

	sm.shapes["bubs"] = parseShape("40,-160,0,99|-20,-80,0,101|50,0,0,101|-30,-60,0,96", 99)

	sm.shapes["man"] = parseShape("0,210,0,110|30,230,-60,11|-30,230,-60,11|0,200,-70,12|0,120,0,108|0,30,0,110|0,-60,0,110|0,-150,0,110|40,-190,-60,107|-40,-190,-60,107|0,-170,-70,105|0,-190,-80,105|0,-230,-85,105|90,-190,0,108|110,-240,0,108|120,-300,0,108|120,-360,0,108|120,-450,0,110|-90,-190,0,108|-110,-240,0,108|-120,-300,0,108|-120,-360,0,108|-120,-450,0,110|100,30,0,108|130,-20,20,108|150,-70,40,108|160,-130,40,108|-100,30,0,108|-130,-20,20,108|-150,-70,40,108|-160,-130,40,108|160,-140,0,107|150,-145,-40,106|145,-150,-80,105|-160,-140,0,107|-150,-145,-40,106|-145,-150,-80,105", 110)

	sm.shapes["woman"] = parseShape("0,210,0,110|30,230,-60,83|-30,230,-60,83|0,200,-70,84|0,120,0,108|0,30,0,110|0,-60,0,110|0,-150,0,110|50,0,-70,110|-50,0,-70,110|50,0,-130,80|-50,0,-130,80|40,-190,-50,82|-40,-190,-50,82|0,-230,-40,82|90,-190,0,108|110,-240,0,108|120,-300,0,108|120,-360,0,108|120,-450,0,110|-90,-190,0,108|-110,-240,0,108|-120,-300,0,108|-120,-360,0,108|-120,-450,0,110|100,30,0,108|130,-20,20,108|150,-70,40,108|160,-130,40,108|-100,30,0,108|-130,-20,20,108|-150,-70,40,108|-160,-130,40,108|160,-140,0,107|150,-145,-40,106|145,-150,-80,105|-160,-140,0,107|-150,-145,-40,106|-145,-150,-80,105", 110)

	sm.shapes["tardi"] = parseShape("0,280,0,5|-40,200,-40,93|-40,200,40,93|40,200,-40,93|40,200,40,93|-80,120,-80,93|-80,120,0,93|-80,120,80,93|80,120,-80,93|80,120,0,93|80,120,80,93|0,120,-80,93|0,120,80,93|-70,40,-70,92|-70,40,0,92|-70,40,70,92|70,40,-70,92|70,40,0,92|70,40,70,92|0,40,-70,92|0,40,70,92|-70,-40,-70,92|-70,-40,0,20|-70,-40,70,92|70,-40,-70,92|70,-40,0,20|70,-40,70,92|0,-40,-70,20|0,-40,70,20|-70,-120,-70,92|-70,-120,0,92|-70,-120,70,92|70,-120,-70,92|70,-120,0,92|70,-120,70,92|0,-120,-70,92|0,-120,70,92|-70,-200,-70,92|-70,-200,0,92|-70,-200,70,92|70,-200,-70,92|70,-200,0,92|70,-200,70,92|0,-200,-70,92|0,-200,70,92|-70,-280,-70,92|-70,-280,0,92|-70,-280,70,92|70,-280,-70,92|70,-280,0,92|70,-280,70,92|0,-280,-70,92|0,-280,70,92|0,-280,0,92", 5)

	sm.shapes["spider"] = parseShape("0,0,-80,118|0,0,80,118|40,13,150,5|-40,13,150,5|120,0,-108,117|165,60,-118,117|210,0,-138,117|255,-60,-158,117|288,-120,-178,117|-120,0,-108,117|-165,60,-118,117|-210,0,-138,117|-255,-60,-158,117|-288,-120,-178,117|120,0,70,117|165,60,93,117|210,0,100,117|255,-60,120,117|288,-120,140,117|-120,0,70,117|-165,60,93,117|-210,0,100,117|-255,-60,120,117|-288,-120,140,117|120,0,12,117|165,60,12,117|210,0,12,117|255,-60,12,117|288,-120,12,117|-120,0,12,117|-165,60,12,117|-210,0,12,117|-255,-60,12,117|-288,-120,12,117|120,0,-63,117|165,60,-63,117|210,0,-63,117|255,-60,-63,117|288,-120,-63,117|-120,0,-63,117|-165,60,-63,117|-210,0,-63,117|-255,-60,-63,117|-288,-120,-63,117|0,60,0,16|0,140,0,16|0,220,0,16|0,300,0,16|0,380,0,16|0,460,0,16|0,540,0,16|0,620,0,16|0,700,0,16|0,780,0,16|0,860,0,16|0,940,0,16|0,1020,0,16|0,1100,0,16", 118)

	sm.shapes["cube"] = parseShape("140,140,140,120|0,140,140,120|-140,140,140,120|140,0,140,120|0,0,140,120|-140,0,140,120|140,-140,140,120|0,-140,140,120|-140,-140,140,120|140,140,0,120|0,140,0,120|-140,140,0,120|140,0,0,120|0,0,0,120|-140,0,0,120|140,-140,0,120|0,-140,0,120|-140,-140,0,120|140,140,-140,120|0,140,-140,120|-140,140,-140,120|140,0,-140,120|0,0,-140,120|-140,0,-140,120|140,-140,-140,120|0,-140,-140,120|-140,-140,-140,120", 120)

	sm.shapes["spacecube"] = parseShape("140,140,0,120|0,140,0,120|-140,140,0,120|140,0,0,120|0,0,0,120|-140,0,0,120|140,-140,0,120|0,-140,0,120|-140,-140,0,120|0,0,140,120|0,0,-140,120|-280,0,0,120|280,0,0,120|-420,0,0,120|420,0,0,120|0,0,280,120|0,0,-280,120|0,-280,0,120|0,-420,0,120|0,280,0,120|0,420,0,120|140,420,0,120|-140,420,0,120|140,-420,0,120|-140,-420,0,120|0,0,420,120|0,0,-420,120", 120)
}

// GetCopy returns a copy of the shape
func (sm *ShapeManager) GetCopy(name string) *Shape {
	original, ok := sm.shapes[name]
	if !ok {
		return &Shape{Points: []Vector3{}}
	}

	// Deep copy
	points := make([]Vector3, len(original.Points))
	copy(points, original.Points)
	return &Shape{Points: points}
}

// YMPlayer wraps the YM player for use with Ebiten's audio system
type YMPlayer struct {
	player     *stsound.StSound
	sampleRate int
	buffer     []int16
	mutex      sync.Mutex
	position   int64 // PCM byte offset
	totalBytes int64
	loop       bool
}

// NewYMPlayer creates a new YM player instance
func NewYMPlayer(data []byte, sampleRate int, loop bool) (*YMPlayer, error) {
	player := stsound.CreateWithRate(sampleRate)

	if err := player.LoadMemory(data); err != nil {
		player.Destroy()
		return nil, fmt.Errorf("failed to load YM data: %w", err)
	}

	player.SetLoopMode(loop)

	info := player.GetInfo()
	totalSamples := int64(info.MusicTimeInMs) * int64(sampleRate) / 1000

	return &YMPlayer{
		player:     player,
		sampleRate: sampleRate,
		buffer:     make([]int16, 4096),
		totalBytes: totalSamples * 4,
		loop:       loop,
	}, nil
}

// Read implements io.Reader for audio streaming
func (y *YMPlayer) Read(p []byte) (n int, err error) {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	// Each frame is one signed 16-bit sample duplicated to two channels.
	byteCount := len(p) &^ 3
	samplesNeeded := byteCount / 4

	processed := 0
	for processed < samplesNeeded {
		chunkSize := min(samplesNeeded-processed, len(y.buffer))

		if !y.player.Compute(y.buffer[:chunkSize], chunkSize) {
			if !y.loop {
				clear(p[processed*4 : byteCount])
				y.position = y.totalBytes
				return byteCount, io.EOF
			}
		}

		for i := 0; i < chunkSize; i++ {
			sample := y.buffer[i]
			offset := (processed + i) * 4
			p[offset] = byte(sample)
			p[offset+1] = byte(sample >> 8)
			p[offset+2] = byte(sample)
			p[offset+3] = byte(sample >> 8)
		}

		processed += chunkSize
		y.position += int64(chunkSize * 4)
	}

	return byteCount, nil
}

// Seek implements io.Seeker
func (y *YMPlayer) Seek(offset int64, whence int) (int64, error) {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = y.position + offset
	case io.SeekEnd:
		newPos = y.totalBytes + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newPos < 0 {
		newPos = 0
	}
	if newPos > y.totalBytes {
		newPos = y.totalBytes
	}
	newPos &^= 3

	y.player.Seek(uint32(newPos / 4 * 1000 / int64(y.sampleRate)))
	y.position = newPos
	return newPos, nil
}

// Close releases resources
func (y *YMPlayer) Close() error {
	y.mutex.Lock()
	defer y.mutex.Unlock()

	if y.player != nil {
		y.player.Destroy()
		y.player = nil
	}
	return nil
}

// Length returns total length
func (y *YMPlayer) Length() int64 {
	return y.totalBytes
}

// Animation interface
type Animation interface {
	Run(points []Vector3, frameCount int)
}

// Sinus2D animation
type Sinus2D struct {
	ctr    float64
	ctrAmp float64
}

func (s *Sinus2D) Run(points []Vector3, frameCount int) {
	amplitude := 200 * math.Sin(s.ctrAmp)
	sinCtr, cosCtr := math.Sincos(s.ctr)
	i := 0
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if i < len(points) {
				phase := x + y
				points[i].Z = amplitude * (cosCtr*sinusPhaseCos[phase] - sinCtr*sinusPhaseSin[phase])
			}
			i++
		}
	}
	s.ctr += math.Pi / 55
	s.ctrAmp += math.Pi / 60
}

// MorphingTo animation
type MorphingTo struct {
	steps      []Vector3
	frameCount int
	maxFrames  int
}

func NewMorphingTo(from, to []Vector3, nbFrames int) *MorphingTo {
	steps := make([]Vector3, len(from))
	for i := range from {
		if i < len(to) {
			steps[i] = Vector3{
				X:   (to[i].X - from[i].X) / float64(nbFrames),
				Y:   (to[i].Y - from[i].Y) / float64(nbFrames),
				Z:   (to[i].Z - from[i].Z) / float64(nbFrames),
				Img: 0,
			}
		}
	}
	return &MorphingTo{
		steps:     steps,
		maxFrames: nbFrames,
	}
}

func (m *MorphingTo) Run(points []Vector3, frameCount int) {
	if m.frameCount >= m.maxFrames {
		return
	}
	for i := range points {
		if i < len(m.steps) {
			points[i].X += m.steps[i].X
			points[i].Y += m.steps[i].Y
			points[i].Z += m.steps[i].Z
		}
	}
	m.frameCount++
}

// YRotate animation
type YRotate struct {
	ctr   float64
	ratio float64
	incr  float64
}

func NewYRotate() *YRotate {
	return &YRotate{
		ctr:   0,
		ratio: 0,
		incr:  0.025,
	}
}

func (y *YRotate) Run(points []Vector3, frameCount int) {
	// This animation modifies position via GetPosition()
	y.ctr += math.Pi / 144
	if y.incr > 0 {
		y.ratio += y.incr
		if y.ratio >= 1 {
			y.incr = 0
		}
	} else if y.incr < 0 {
		y.ratio += y.incr
		if y.ratio <= 0 {
			y.incr = 0
		}
	}
}

// GetPosition returns the new position for YRotate animation
func (y *YRotate) GetPosition() Vector3 {
	sinCtr, cosCtr := math.Sincos(y.ctr)
	return Vector3{
		X: y.ratio * 100 * cosCtr,
		Y: 0,
		Z: 850 + y.ratio*100*sinCtr,
	}
}

// Rotors animation (for helicopter)
type Rotors struct {
	ctr float64
}

func (r *Rotors) Run(points []Vector3, frameCount int) {
	if len(points) < 8 {
		return
	}

	// DON'T apply scaleFactor here - the points are already scaled from parseShape
	radii := [...]float64{80, 150, 210, 160}
	offset := 100.0
	sinCtr, cosCtr := math.Sincos(r.ctr)

	for i, radius := range radii {
		x := radius * sinCtr
		z := radius * cosCtr
		if i*2 < len(points) {
			points[i*2].X = x
			// DON'T touch Y - it's already set correctly from the shape
			points[i*2].Z = z - offset
		}
		if i*2+1 < len(points) {
			points[i*2+1].X = -x
			// DON'T touch Y - it's already set correctly from the shape
			points[i*2+1].Z = -z - offset
		}
	}
	r.ctr += 0.08
}

// Bounce animation
type Bounce struct {
	curve      []float64
	ctr        int
	bounceOffY float64
}

func NewBounce() *Bounce {
	curve := make([]float64, 0, 540)
	rad := 1200.0
	off := 210.0

	formula := func(a float64) {
		curve = append(curve, -(off-rad*math.Cos(a*math.Pi/180))/3)
		rad -= 5.7143 / frameRateConversion
		off -= 1 / frameRateConversion
	}

	for c := 1; c < 4; c++ {
		for a := 0.0; a < 90; a += 3 / frameRateConversion {
			formula(a)
		}
		for a := 89.0; a >= 0; a -= 3 / frameRateConversion {
			formula(a)
		}
	}
	for a := 0.0; a < 90; a++ {
		formula(a)
	}

	return &Bounce{curve: curve, ctr: 0, bounceOffY: 0}
}

func (b *Bounce) Run(points []Vector3, frameCount int) {
	// Update bounce offset - this will be applied to position.Y
	if b.ctr < len(b.curve) {
		b.bounceOffY = b.curve[b.ctr]
	}
	b.ctr++
	if b.ctr >= len(b.curve) {
		b.ctr = 0
	}
}

// GetBounceY returns the current bounce Y position (not offset!)
func (b *Bounce) GetBounceY() float64 {
	return b.bounceOffY
}

// Action represents a timeline action
type Action struct {
	Shape        string
	Pos          *Vector3
	InitRot      *Vector3
	Rot          *Vector3
	Tr           *Vector3
	AnimTypes    []string
	Frames       int
	TextIndex    int
	InitFunc     func(*Game)
	hasText      bool
	hasShape     bool
	hasPos       bool
	hasInitRot   bool
	hasRot       bool
	hasTr        bool
	hasAnimTypes bool // true if AnimTypes was explicitly set (even if empty)
}

// Game represents the main demo state
type Game struct {
	sharedProjector sprites.Projector
	sharedPoints    []sprites.Point
	// Images
	ballsSource image.Image
	ballsAtlas  *ebiten.Image
	textImg     *ebiten.Image
	textRows    []*ebiten.Image
	whiteImage  *ebiten.Image

	// Ball sprites (extracted from AllBalls.png)
	balls []*ebiten.Image

	// Canvases
	playgroundCanvas *ebiten.Image
	reflectionSource *ebiten.Image

	// 3D state
	shapeManager *ShapeManager
	currentShape *Shape
	position     Vector3
	rotation     Vector3
	rotSpeed     Vector3
	trSpeed      Vector3
	zoomFactor   float64
	transformed  []Point3D

	// Animation
	animations []Animation
	frameCount int

	// Actions timeline
	actions       []Action
	currentAction int
	actionFrames  int

	// Audio
	audioContext *audio.Context
	audioPlayer  *audio.Player
	ymPlayer     *YMPlayer

	// Text
	currentText int

	// Projection
	fov     float64
	centerX float64
	centerY float64

	// Initialization
	ballsExtracted bool
	dirty          bool
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		shapeManager: NewShapeManager(),
		zoomFactor:   0.35,
		fov:          600 + 850,
		centerX:      320,
		centerY:      193,
		position:     Vector3{X: 0, Y: 0, Z: 850, Img: 0},
		transformed:  make([]Point3D, 0, 64),
		dirty:        true,
	}

	// Load images
	g.loadImages()

	// Create canvases
	g.playgroundCanvas = ebiten.NewImage(640, 386)
	g.reflectionSource = g.playgroundCanvas.SubImage(image.Rect(0, 288, 640, 368)).(*ebiten.Image)
	g.whiteImage = ebiten.NewImage(1, 1)
	g.whiteImage.Fill(color.White)

	// Initialize actions timeline
	g.initActions()

	// Initialize audio
	g.initAudio()

	// Start first action
	g.currentAction = -1
	g.nextAction()

	return g
}

// loadImages loads all image assets
func (g *Game) loadImages() {
	var err error

	// Load balls image
	img, _, err := image.Decode(bytes.NewReader(ballsData))
	if err != nil {
		log.Printf("Failed to load balls image: %v", err)
	} else {
		g.ballsSource = img
	}

	// Load text image
	img, _, err = image.Decode(bytes.NewReader(textData))
	if err != nil {
		log.Printf("Failed to load text image: %v", err)
	} else {
		g.textImg = ebiten.NewImageFromImage(img)
		for y := 0; y+14 <= img.Bounds().Dy(); y += 14 {
			row := g.textImg.SubImage(image.Rect(0, y, 640, y+14)).(*ebiten.Image)
			g.textRows = append(g.textRows, row)
		}
	}
}

// extractBalls creates all ball sprite variations
func (g *Game) extractBalls() {
	if g.ballsSource == nil {
		return
	}

	// Color palettes (RGB values 0-255)
	palettes := [][]color.RGBA{
		{{108, 0, 0, 255}, {144, 0, 0, 255}, {180, 0, 0, 255}, {216, 0, 0, 255}, {252, 0, 0, 255}},
		{{0, 108, 108, 255}, {0, 144, 144, 255}, {0, 180, 180, 255}, {0, 216, 216, 255}, {0, 252, 252, 255}},
		{{108, 108, 108, 255}, {144, 144, 144, 255}, {180, 180, 180, 255}, {216, 216, 216, 255}, {252, 252, 252, 255}},
		{{108, 0, 72, 255}, {144, 0, 108, 255}, {180, 0, 144, 255}, {216, 0, 180, 255}, {252, 0, 216, 255}},
		{{72, 108, 0, 255}, {108, 144, 0, 255}, {144, 180, 0, 255}, {180, 216, 0, 255}, {216, 252, 0, 255}},
		{{108, 36, 36, 255}, {144, 36, 72, 255}, {180, 36, 108, 255}, {216, 36, 144, 255}, {252, 36, 180, 255}},
		{{108, 72, 0, 255}, {108, 72, 36, 255}, {180, 72, 72, 255}, {216, 72, 108, 255}, {252, 72, 144, 255}},
		{{72, 108, 0, 255}, {108, 108, 0, 255}, {144, 108, 36, 255}, {180, 108, 72, 255}, {252, 108, 108, 255}},
		{{72, 108, 0, 255}, {108, 144, 0, 255}, {144, 144, 0, 255}, {180, 144, 36, 255}, {216, 180, 36, 255}},
		{{0, 108, 0, 255}, {0, 144, 0, 255}, {0, 180, 0, 255}, {0, 216, 0, 255}, {0, 252, 0, 255}},
		{{108, 72, 0, 255}, {144, 108, 0, 255}, {180, 144, 0, 255}, {216, 180, 0, 255}, {252, 216, 0, 255}},
		{{0, 36, 108, 255}, {0, 36, 144, 255}, {0, 36, 180, 255}, {0, 72, 216, 255}, {0, 72, 252, 255}},
		{{0, 72, 108, 255}, {0, 108, 144, 255}, {0, 144, 180, 255}, {0, 180, 216, 255}, {0, 216, 252, 255}},
		{{180, 72, 72, 255}, {180, 108, 108, 255}, {216, 144, 144, 255}, {252, 180, 180, 255}, {252, 216, 216, 255}},
		{{36, 108, 0, 255}, {72, 144, 0, 255}, {108, 180, 0, 255}, {144, 216, 0, 255}, {180, 252, 0, 255}},
	}

	sizes := [...]int{12, 16, 20, 24, 28, 32, 54}
	atlasWidth := 0
	for _, size := range sizes {
		atlasWidth += size
	}
	const rowHeight = 54
	atlas := image.NewRGBA(image.Rect(0, 0, atlasWidth, (len(palettes)+1)*rowHeight))

	for paletteIndex, palette := range palettes {
		sourceY := 0
		destinationX := 0
		for _, size := range sizes {
			sourceRect := image.Rect(0, sourceY, size, sourceY+size)
			destination := image.Pt(destinationX, paletteIndex*rowHeight)
			recolorBall(atlas, destination, g.ballsSource, sourceRect, palette)
			sourceY += size
			destinationX += size
		}
	}

	checkedSource := image.Rect(0, 186, 54, 240)
	checkedDestination := image.Pt(0, len(palettes)*rowHeight)
	copyImage(atlas, checkedDestination, g.ballsSource, checkedSource)

	g.ballsAtlas = ebiten.NewImageFromImage(atlas)
	g.balls = make([]*ebiten.Image, 0, len(palettes)*(len(sizes)+1)+1)
	for paletteIndex := range palettes {
		x := 0
		for _, size := range sizes {
			rect := image.Rect(x, paletteIndex*rowHeight, x+size, paletteIndex*rowHeight+size)
			g.balls = append(g.balls, g.ballsAtlas.SubImage(rect).(*ebiten.Image))
			x += size
		}
		// The source demo repeats the last ball of each palette.
		g.balls = append(g.balls, g.balls[len(g.balls)-1])
	}
	checkedRect := image.Rect(0, len(palettes)*rowHeight, 54, (len(palettes)+1)*rowHeight)
	g.balls = append(g.balls, g.ballsAtlas.SubImage(checkedRect).(*ebiten.Image))
	g.ballsSource = nil
}

// recolorBall writes a recolored sprite into a CPU-side atlas. Keeping this
// work off ebiten.Image avoids synchronous GPU readbacks during startup.
func recolorBall(dst *image.RGBA, destination image.Point, src image.Image, source image.Rectangle, palette []color.RGBA) {
	for y := 0; y < source.Dy(); y++ {
		for x := 0; x < source.Dx(); x++ {
			c := src.At(source.Min.X+x, source.Min.Y+y)
			r, _, _, a := c.RGBA()
			if a > 0 {
				redVal := uint8(r >> 8)
				index := int(redVal>>5) - 3
				if index >= 0 && index < len(palette) {
					dst.Set(destination.X+x, destination.Y+y, palette[index])
				} else {
					dst.Set(destination.X+x, destination.Y+y, c)
				}
			}
		}
	}
}

func copyImage(dst *image.RGBA, destination image.Point, src image.Image, source image.Rectangle) {
	for y := 0; y < source.Dy(); y++ {
		for x := 0; x < source.Dx(); x++ {
			dst.Set(destination.X+x, destination.Y+y, src.At(source.Min.X+x, source.Min.Y+y))
		}
	}
}

// initAudio initializes the audio system with YM music
func (g *Game) initAudio() {
	g.audioContext = audio.NewContext(audioSampleRate)

	var err error
	g.ymPlayer, err = NewYMPlayer(musicData, audioSampleRate, true)
	if err != nil {
		log.Printf("Failed to create YM player: %v", err)
		return
	}

	g.audioPlayer, err = g.audioContext.NewPlayer(g.ymPlayer)
	if err != nil {
		log.Printf("Failed to create audio player: %v", err)
		g.ymPlayer.Close()
		g.ymPlayer = nil
		return
	}

	g.audioPlayer.SetVolume(0.7)
	g.audioPlayer.Play()
}

// initActions initializes the action timeline
func (g *Game) initActions() {
	toRad := math.Pi / 180
	toRadSpeed := toRad / frameRateConversion

	g.actions = []Action{
		// SQUARE
		{Shape: "square", Pos: &Vector3{X: 0, Y: 320, Z: 850}, InitRot: &Vector3{X: 30 * toRad, Y: 30 * toRad, Z: 30 * toRad},
			Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, AnimTypes: []string{}, Frames: 45, TextIndex: 0,
			hasShape: true, hasPos: true, hasInitRot: true, hasRot: true, hasTr: true, hasAnimTypes: true, hasText: true},
		{Pos: &Vector3{X: 0, Y: 0, Z: 850}, Tr: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 45, hasPos: true, hasTr: true},
		{Frames: 30, TextIndex: 1, hasText: true},

		// RIPP_IN + RIPPLE
		{AnimTypes: []string{"Sinus2D", "YRotate"}, Frames: 60, hasAnimTypes: true},
		{Frames: 60, TextIndex: 2, hasText: true},
		{Frames: 60, TextIndex: 3, hasText: true},
		{Frames: 120},

		// MERGE to sphere
		{AnimTypes: []string{"MorphingSphere"}, Frames: 45, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 4 * toRadSpeed}, hasRot: true, hasAnimTypes: true},
		{AnimTypes: []string{}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 80, TextIndex: 4, hasRot: true, hasAnimTypes: true, hasText: true},
		{Frames: 40, TextIndex: 5, hasText: true},
		{AnimTypes: []string{"YRotate"}, Frames: 240, hasAnimTypes: true},

		// TUBE
		{AnimTypes: []string{"MorphingTube"}, Frames: 45, TextIndex: 6, hasText: true, hasAnimTypes: true},
		{Frames: 60},
		{Shape: "tube", Frames: 60, TextIndex: 7, hasShape: true, hasText: true},
		{Frames: 80},
		{AnimTypes: []string{"MorphingSquare"}, Frames: 60, TextIndex: 8, hasText: true, hasAnimTypes: true},
		{AnimTypes: []string{}, Frames: 50, TextIndex: 9, hasText: true, hasAnimTypes: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Frames: 50, hasTr: true},

		// HELI
		{Shape: "heli", AnimTypes: []string{"Rotors"}, Pos: &Vector3{X: 0, Y: 320, Z: 850}, InitRot: &Vector3{X: 180 * toRad, Y: 0, Z: 0},
			Tr: &Vector3{X: 0, Y: -3, Z: 0}, Rot: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 45, TextIndex: 10,
			hasShape: true, hasPos: true, hasInitRot: true, hasTr: true, hasRot: true, hasAnimTypes: true, hasText: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 60, hasTr: true},
		{Rot: &Vector3{X: 0, Y: -3 * toRadSpeed, Z: 0}, Frames: 30, hasRot: true},
		{Frames: 60, TextIndex: 11, hasText: true},
		{Rot: &Vector3{X: toRadSpeed, Y: 0, Z: toRadSpeed}, Frames: 20, hasRot: true},
		{InitFunc: func(g *Game) { g.animations = append(g.animations, NewYRotate()) }, Rot: &Vector3{X: 0, Y: math.Pi * toRadSpeed, Z: 0}, Frames: 240, hasRot: true},
		{Rot: &Vector3{X: math.Pi * toRadSpeed, Y: math.Pi * toRadSpeed, Z: 0}, Frames: 220, TextIndex: 12, hasRot: true, hasText: true},
		{InitFunc: func(g *Game) {
			if len(g.animations) > 0 {
				g.animations = g.animations[:len(g.animations)-1]
			}
		}, Rot: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 20, hasRot: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Frames: 45, hasTr: true},

		// ANIMAL - IMPORTANT: anim:[] clears the Rotors animation from helicopter!
		{Shape: "animal", AnimTypes: []string{}, Pos: &Vector3{X: 0, Y: 320, Z: 850}, InitRot: &Vector3{X: 180 * toRad, Y: 90 * toRad, Z: 0},
			Tr: &Vector3{X: 0, Y: -3, Z: 0}, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 0}, Frames: 45, TextIndex: 13,
			hasShape: true, hasPos: true, hasInitRot: true, hasTr: true, hasRot: true, hasAnimTypes: true, hasText: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 0, Y: 3 * toRadSpeed, Z: 0}, Frames: 180, hasTr: true, hasRot: true},
		{Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 120, TextIndex: 14, hasRot: true, hasText: true},
		{Tr: &Vector3{X: 0, Y: -1.25, Z: 0}, Frames: 120, hasTr: true},

		// BUBS
		{Shape: "bubs", Pos: &Vector3{X: 0, Y: -320, Z: 850}, Rot: &Vector3{X: 0, Y: 5 * toRadSpeed, Z: 0}, Tr: &Vector3{X: 0, Y: 1.5, Z: 0}, Frames: 240, TextIndex: 15,
			hasShape: true, hasPos: true, hasRot: true, hasTr: true, hasText: true},

		// MAN
		{Shape: "man", Pos: &Vector3{X: 0, Y: 320, Z: 850}, Rot: &Vector3{X: 0, Y: 4 * toRadSpeed, Z: 0}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, Frames: 45, TextIndex: 16,
			hasShape: true, hasPos: true, hasRot: true, hasTr: true, hasText: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 2 * toRadSpeed, Y: 4 * toRadSpeed, Z: 2 * toRadSpeed}, Frames: 240, hasTr: true, hasRot: true},
		{Frames: 120, TextIndex: 17, hasText: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Frames: 45, hasTr: true},

		// WOMAN
		{Shape: "woman", Pos: &Vector3{X: 0, Y: 320, Z: 850}, Rot: &Vector3{X: -2 * toRadSpeed, Y: 4 * toRadSpeed, Z: -2 * toRadSpeed}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, Frames: 45,
			hasShape: true, hasPos: true, hasRot: true, hasTr: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 0, Y: 4 * toRadSpeed, Z: 2 * toRadSpeed}, Frames: 360, TextIndex: 18, hasTr: true, hasRot: true, hasText: true},
		{Tr: &Vector3{X: 0, Y: -3, Z: 0}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 4 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 45, hasTr: true, hasRot: true},

		// TARDI transitions
		{Shape: "tardi", Pos: &Vector3{X: 0, Y: 0, Z: 850}, InitRot: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 0, Y: 3 * toRadSpeed, Z: 0}, Tr: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 20, TextIndex: 19,
			hasShape: true, hasPos: true, hasInitRot: true, hasRot: true, hasTr: true, hasText: true},
		{Frames: 10}, {Frames: 10}, {Frames: 10}, {Frames: 10}, {Frames: 10}, {Frames: 10},

		// TARDI
		{Shape: "tardi", Frames: 90, hasShape: true},
		{Frames: 90, TextIndex: 20, hasText: true},
		{Rot: &Vector3{X: -3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 0}, AnimTypes: []string{"YRotate"}, Frames: 120, TextIndex: 21, hasRot: true, hasAnimTypes: true, hasText: true},
		{Tr: &Vector3{X: 0, Y: -2, Z: 0}, Frames: 120, TextIndex: 22, hasTr: true, hasText: true},

		// SPIDER
		{Shape: "spider", Tr: &Vector3{X: 0, Y: 0, Z: 0}, Pos: &Vector3{X: 0, Y: -900, Z: 850}, InitRot: &Vector3{X: 0, Y: 180 * toRad, Z: 0},
			Rot: &Vector3{X: 0, Y: 3 * toRadSpeed, Z: 0}, AnimTypes: []string{"Bounce"}, Frames: 215, TextIndex: 23,
			hasShape: true, hasTr: true, hasPos: true, hasInitRot: true, hasRot: true, hasAnimTypes: true, hasText: true},
		{AnimTypes: []string{}, Frames: 75, TextIndex: 24, hasAnimTypes: true, hasText: true},
		{Frames: 75, TextIndex: 25, hasText: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Rot: &Vector3{X: 0, Y: 2 * toRadSpeed, Z: 0}, Frames: 45, TextIndex: 26, hasTr: true, hasRot: true, hasText: true},

		// CUBE
		{Shape: "cube", Pos: &Vector3{X: 0, Y: 320, Z: 850}, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 4 * toRadSpeed}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, Frames: 45,
			hasShape: true, hasPos: true, hasRot: true, hasTr: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 240, TextIndex: 27, hasTr: true, hasRot: true, hasText: true},
		{AnimTypes: []string{"MorphingSpaceCube"}, Frames: 45, hasAnimTypes: true},
		{AnimTypes: []string{}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 120, hasRot: true, hasAnimTypes: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 4 * toRadSpeed}, Frames: 45, hasTr: true, hasRot: true},
	}
}

// nextAction advances to the next action in the timeline
func (g *Game) nextAction() {
	g.currentAction++
	if g.currentAction >= len(g.actions) {
		g.currentAction = 0
	}

	action := &g.actions[g.currentAction]
	g.actionFrames = int(float64(action.Frames) * frameRateConversion)

	// Apply action properties
	if action.hasShape {
		g.currentShape = g.shapeManager.GetCopy(action.Shape)
	}
	if action.hasPos {
		g.position = *action.Pos
	}
	if action.hasInitRot {
		g.rotation = *action.InitRot
	}
	if action.hasRot {
		g.rotSpeed = *action.Rot
	}
	if action.hasTr {
		g.trSpeed = *action.Tr
	}
	if action.hasText {
		g.currentText = action.TextIndex
	}

	// Initialize animations - ONLY if AnimTypes is provided
	// In JS line 184: if(t.anim)  this.anim = t.anim.map(...)
	// When anim is NOT specified (hasAnimTypes=false), keep the previous animations!
	// When anim IS specified but empty (hasAnimTypes=true, len=0), clear animations!
	if action.hasAnimTypes {
		g.animations = g.animations[:0]
		for _, animType := range action.AnimTypes {
			switch animType {
			case "Sinus2D":
				g.animations = append(g.animations, &Sinus2D{})
			case "YRotate":
				g.animations = append(g.animations, NewYRotate())
			case "Rotors":
				g.animations = append(g.animations, &Rotors{})
			case "Bounce":
				g.animations = append(g.animations, NewBounce())
			case "MorphingSphere":
				target := g.shapeManager.shapes["sphere"]
				g.animations = append(g.animations, NewMorphingTo(g.currentShape.Points, target.Points, int(45*frameRateConversion)))
			case "MorphingTube":
				target := g.shapeManager.shapes["tube"]
				g.animations = append(g.animations, NewMorphingTo(g.currentShape.Points, target.Points, int(45*frameRateConversion)))
			case "MorphingSquare":
				target := g.shapeManager.shapes["square"]
				g.animations = append(g.animations, NewMorphingTo(g.currentShape.Points, target.Points, int(60*frameRateConversion)))
			case "MorphingSpaceCube":
				target := g.shapeManager.shapes["spacecube"]
				g.animations = append(g.animations, NewMorphingTo(g.currentShape.Points, target.Points, int(45*frameRateConversion)))
			}
		}
	}

	if action.InitFunc != nil {
		action.InitFunc(g)
	}
}

// Update updates the game state
func (g *Game) Update() error {
	g.dirty = true

	// Extract balls on first frame (after game starts)
	if !g.ballsExtracted {
		g.extractBalls()
		g.ballsExtracted = true
	}

	g.frameCount++

	// Update action timeline
	g.actionFrames--
	if g.actionFrames <= 0 {
		g.nextAction()
	} else {
		// Apply rotation
		g.rotation.X += g.rotSpeed.X
		g.rotation.Y += g.rotSpeed.Y
		g.rotation.Z += g.rotSpeed.Z

		// Apply translation (BEFORE animations, just like in JS lines 201-203)
		g.position.X += g.trSpeed.X
		g.position.Y += g.trSpeed.Y
		g.position.Z += g.trSpeed.Z

		// Run animations (they can override position)
		// In JS line 205: this.anim.forEach(a => a.run(this.the3d));
		if g.currentShape != nil {
			for _, anim := range g.animations {
				anim.Run(g.currentShape.Points, g.frameCount)
				// Apply bounce Y position if Bounce animation is active
				// Note: Bounce REPLACES position.Y, it doesn't add to it
				if bounce, ok := anim.(*Bounce); ok {
					g.position.Y = bounce.GetBounceY()
				}
				// Apply YRotate position if active
				// Note: YRotate REPLACES the entire position, overriding the trSpeed above
				if yrot, ok := anim.(*YRotate); ok {
					g.position = yrot.GetPosition()
				}
			}
		}
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.dirty {
		return
	}
	g.dirty = false

	screen.Clear()
	g.playgroundCanvas.Clear()

	// Draw 3D scene
	g.draw3D()

	// Draw text tile (if available)
	if g.currentText >= 0 && g.currentText < len(g.textRows) {
		screen.DrawImage(g.textRows[g.currentText], &ebiten.DrawImageOptions{})
	}

	// Draw playground canvas
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(0, 14)
	screen.DrawImage(g.playgroundCanvas, opts)

	// Draw blue lines
	g.drawBlueLines(screen)

	// Draw reflection
	g.drawReflection(screen)
}

// draw3D renders the 3D vectorball scene
func (g *Game) draw3D() {
	if g.currentShape == nil || len(g.currentShape.Points) == 0 {
		return
	}
	if cap(g.sharedPoints) < len(g.currentShape.Points) {
		g.sharedPoints = make([]sprites.Point, len(g.currentShape.Points))
	} else {
		g.sharedPoints = g.sharedPoints[:len(g.currentShape.Points)]
	}
	for i, p := range g.currentShape.Points {
		g.sharedPoints[i] = sprites.Point{X: p.X, Y: p.Y, Z: p.Z, Image: p.Img}
	}
	g.sharedProjector.Draw(g.playgroundCanvas, g.sharedPoints, g.balls, sprites.Projection{Matrix: [9]float64(newRotationMatrix(g.rotation, g.zoomFactor)), Translate: sprites.Point{X: g.position.X, Y: g.position.Y, Z: g.position.Z}, Focal: g.fov, CenterX: g.centerX, CenterY: g.centerY, YUp: true, AscendingDepth: true})
}

// drawBlueLines draws the horizontal blue separator lines
func (g *Game) drawBlueLines(screen *ebiten.Image) {
	// Horizontal lines
	drawRect(screen, g.whiteImage, 0, 382, 640, 18, color.RGBA{0, 0, 68, 255})
	drawRect(screen, g.whiteImage, 0, 378, 640, 4, color.RGBA{0, 0, 34, 255})
	drawRect(screen, g.whiteImage, 0, 386, 640, 2, color.RGBA{0, 0, 34, 255})
	drawRect(screen, g.whiteImage, 0, 394, 640, 2, color.RGBA{0, 0, 88, 255})
}

// drawReflection draws the reflection effect at the bottom
func (g *Game) drawReflection(screen *ebiten.Image) {
	// Blue background for reflection area
	drawRect(screen, g.whiteImage, 0, 400, 640, 80, color.RGBA{0, 0, 122, 255})

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(1, -1)
	opts.GeoM.Translate(0, 480)
	opts.ColorScale.ScaleAlpha(0.5)
	composite.Instance{Image: g.reflectionSource, Options: *opts}.Draw(screen)
}

// drawRect draws a filled rectangle
func drawRect(dst, white *ebiten.Image, x, y, width, height int, c color.RGBA) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(float64(width), float64(height))
	opts.GeoM.Translate(float64(x), float64(y))
	opts.ColorScale.Scale(
		float32(c.R)/255,
		float32(c.G)/255,
		float32(c.B)/255,
		float32(c.A)/255,
	)
	composite.Instance{Image: white, Options: *opts}.Draw(dst)
}

// Layout returns the screen dimensions
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Cleanup releases resources
func (g *Game) Cleanup() {
	if g.audioPlayer != nil {
		g.audioPlayer.Close()
	}
	if g.ymPlayer != nil {
		g.ymPlayer.Close()
	}
}
