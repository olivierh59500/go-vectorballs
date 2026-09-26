package vectorballs

import (
	"bytes"
	originalassets "go-vectorballs"

	"github.com/olivierh59500/democonstructionkit/sound"

	"image"
	"image/color"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/sprites"
	"go-vectorballs/dck/scene"

	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	audio "github.com/olivierh59500/democonstructionkit/sound/output"
)

const (
	screenWidth  = 640
	screenHeight = 480

	// Demo timings (60 fps web vs 25 fps Atari ST)
	frameRateConversion = scene.FrameRateConversion

	// Android and most current desktop audio devices use 48 kHz natively.
	audioSampleRate = 48_000
)

// Embedded assets
var (
	ballsData = originalassets.DCKAssetBallsData()

	textData = originalassets.DCKAssetTextData()

	musicData = originalassets.DCKAssetMusicData()
)

// Vectorball artwork and action tables remain production data in scene.
type Vector3 = scene.Vector3
type Shape = scene.Shape
type ShapeManager = scene.ShapeManager
type Action = scene.Action

func NewShapeManager() *ShapeManager { return scene.NewShapeManager() }

// Game represents the main demo state
type Game struct {
	objectRest, objectAnimated []sprites.Point
	objectFlag                 *sprites.Flag
	sharedProjector            sprites.Projector
	sharedPoints               []sprites.Point
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
	waterReflection  *composite.WaterReflection

	// 3D state
	shapeManager *ShapeManager
	currentShape *Shape
	position     Vector3
	rotation     Vector3
	zoomFactor   float64

	frameCount int

	// Authored action data and its reusable DCK controller.
	actions  []Action
	sequence *geometry.PointSequence

	// Audio
	audioContext *audio.Context
	audioPlayer  *audio.Player
	musicStream  *sound.Stream

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

// initActions loads the production's authored stage data.
func (g *Game) initActions() { g.actions = scene.AuthoredActions() }

// NewGame creates a new game instance.
func NewGame() *Game {
	g := &Game{
		shapeManager: NewShapeManager(),
		zoomFactor:   0.35,
		fov:          600 + 850,
		centerX:      320,
		centerY:      193,
		position:     Vector3{X: 0, Y: 0, Z: 850, Img: 0},
		dirty:        true,
	}

	// Load images
	g.loadImages()

	// Create canvases
	g.playgroundCanvas = ebiten.NewImage(640, 386)
	g.initReflection()
	g.whiteImage = ebiten.NewImage(1, 1)
	g.whiteImage.Fill(color.White)

	// Initialize actions timeline
	g.initActions()

	// Initialize audio
	g.initAudio()

	if err := g.bindSequence(); err != nil {
		panic(err)
	}

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

// initAudio opens the soundtrack and starts audio output.
func (g *Game) initAudio() {
	g.audioContext = audio.NewContext(audioSampleRate)

	var err error
	g.musicStream, err = sound.Open("music.ym", musicData, sound.Options{SampleRate: audioSampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 1})
	if err != nil {
		log.Printf("Failed to open music: %v", err)
		return
	}

	g.audioPlayer, err = g.audioContext.NewPlayer(g.musicStream)
	if err != nil {
		log.Printf("Failed to create audio player: %v", err)
		g.musicStream.Close()
		g.musicStream = nil
		return
	}

	g.audioPlayer.SetVolume(0.7)
	g.audioPlayer.Play()
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
	if g.objectRest != nil {
		g.updateObject()
		return nil
	}

	if err := g.sequence.Step(); err != nil {
		return err
	}
	g.syncSequence()

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
	angles := geometry.Vec3{X: g.rotation.X, Y: g.rotation.Y, Z: g.rotation.Z}
	g.sharedProjector.Draw(g.playgroundCanvas, g.sharedPoints, g.balls, sprites.Projection{Matrix: [9]float64(geometry.RotateXYZScaled(angles, g.zoomFactor)), Translate: sprites.Point{X: g.position.X, Y: g.position.Y, Z: g.position.Z}, Focal: g.fov, CenterX: g.centerX, CenterY: g.centerY, YUp: true, AscendingDepth: true})
}

// drawBlueLines draws the horizontal blue separator lines
func (g *Game) drawBlueLines(screen *ebiten.Image) {
	// Horizontal lines
	drawRect(screen, g.whiteImage, 0, 382, 640, 18, color.RGBA{0, 0, 68, 255})
	drawRect(screen, g.whiteImage, 0, 378, 640, 4, color.RGBA{0, 0, 34, 255})
	drawRect(screen, g.whiteImage, 0, 386, 640, 2, color.RGBA{0, 0, 34, 255})
	drawRect(screen, g.whiteImage, 0, 394, 640, 2, color.RGBA{0, 0, 88, 255})
}

// initReflection preserves the original crop, waterline and opacity. The same
// DCK pass can reflect any live layer and optionally add waves, tint and fading.
func (g *Game) initReflection() {
	config := composite.DefaultWaterReflectionConfig()
	config.Source = image.Rect(0, 288, 640, 368)
	config.Horizon = 400
	var err error
	g.waterReflection, err = composite.NewWaterReflection(config)
	if err != nil {
		panic(err)
	}
}

// drawReflection draws the reflection effect at the bottom.
func (g *Game) drawReflection(screen *ebiten.Image) {
	// Blue background for reflection area
	drawRect(screen, g.whiteImage, 0, 400, 640, 80, color.RGBA{0, 0, 122, 255})

	g.waterReflection.Draw(screen, g.playgroundCanvas, kit.Frame{Time: float64(g.frameCount) / 60})
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
	if g.waterReflection != nil {
		_ = g.waterReflection.Close()
	}
	if g.audioPlayer != nil {
		g.audioPlayer.Close()
	}
	if g.musicStream != nil {
		g.musicStream.Close()
	}
}
