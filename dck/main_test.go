package vectorballs

import (
	"io"
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/sound"
)

func TestRotationMatrixMatchesSequentialRotations(t *testing.T) {
	tests := []struct {
		point    Vector3
		rotation Vector3
		scale    float64
	}{
		{point: Vector3{X: 1, Y: 2, Z: 3}, rotation: Vector3{}, scale: 1},
		{point: Vector3{X: -280, Y: 120, Z: 43}, rotation: Vector3{X: 0.2, Y: 0.4, Z: 0.6}, scale: 0.35},
		{point: Vector3{X: 140, Y: -420, Z: 280}, rotation: Vector3{X: math.Pi, Y: math.Pi / 2, Z: -math.Pi / 3}, scale: 1.6},
	}

	for _, test := range tests {
		wantX, wantY, wantZ := sequentialRotation(test.point, test.rotation, test.scale)
		angles := geometry.Vec3{X: test.rotation.X, Y: test.rotation.Y, Z: test.rotation.Z}
		point := geometry.Vec3{X: test.point.X, Y: test.point.Y, Z: test.point.Z}
		rotated := geometry.RotateXYZScaled(angles, test.scale).Apply(point)
		gotX, gotY, gotZ := rotated.X, rotated.Y, rotated.Z
		if !closeEnough(gotX, wantX) || !closeEnough(gotY, wantY) || !closeEnough(gotZ, wantZ) {
			t.Fatalf("rotation mismatch: got (%v, %v, %v), want (%v, %v, %v)", gotX, gotY, gotZ, wantX, wantY, wantZ)
		}
	}
}

func TestMusicStreamSeekUsesPCMByteOffsets(t *testing.T) {
	player, err := sound.Open("music.ym", musicData, sound.Options{SampleRate: audioSampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer player.Close()
	// Preserve the exact byte position, including the middle of a stereo frame.
	const tail = 97
	target := int64(audioSampleRate*4 + 3)
	sequential := make([]byte, target+tail)
	if _, err := io.ReadFull(player, sequential); err != nil {
		t.Fatal(err)
	}
	if got, err := player.Seek(target, io.SeekStart); err != nil || got != target {
		t.Fatalf("Seek = %d, %v", got, err)
	}
	after := make([]byte, tail)
	if _, err := io.ReadFull(player, after); err != nil {
		t.Fatal(err)
	}
	for i, v := range after {
		if v != sequential[int(target)+i] {
			t.Fatalf("seek did not reproduce PCM at byte %d", i)
		}
	}
}

func TestMusicStreamReadDoesNotAllocate(t *testing.T) {
	player, err := sound.Open("music.ym", musicData, sound.Options{SampleRate: audioSampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer player.Close()

	buffer := make([]byte, 8192)
	allocations := testing.AllocsPerRun(100, func() {
		n, err := player.Read(buffer)
		if err != nil {
			t.Fatal(err)
		}
		if n != len(buffer) {
			t.Fatalf("Read returned %d bytes, want %d", n, len(buffer))
		}
	})
	if allocations != 0 {
		t.Fatalf("MusicStream.Read allocated %.2f objects per call, want zero", allocations)
	}
}

func BenchmarkMusicStreamRead(b *testing.B) {
	player, err := sound.Open("music.ym", musicData, sound.Options{SampleRate: audioSampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 1})
	if err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { _ = player.Close() })

	buffer := make([]byte, 8192)
	b.ReportAllocs()
	b.SetBytes(int64(len(buffer)))
	for b.Loop() {
		if _, err := player.Read(buffer); err != nil {
			b.Fatal(err)
		}
	}
}

func sequentialRotation(point, rotation Vector3, scale float64) (x, y, z float64) {
	x = point.X * scale
	y = point.Y * scale
	z = point.Z * scale

	sinX, cosX := math.Sincos(rotation.X)
	y, z = y*cosX-z*sinX, y*sinX+z*cosX
	sinY, cosY := math.Sincos(rotation.Y)
	x, z = x*cosY+z*sinY, -x*sinY+z*cosY
	sinZ, cosZ := math.Sincos(rotation.Z)
	x, y = x*cosZ-y*sinZ, x*sinZ+y*cosZ
	return x, y, z
}

func closeEnough(a, b float64) bool {
	return math.Abs(a-b) <= 1e-10*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}
