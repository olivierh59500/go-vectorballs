package vectorballs

import (
	"io"
	"math"
	"testing"
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
		gotX, gotY, gotZ := newRotationMatrix(test.rotation, test.scale).apply(test.point)
		if !closeEnough(gotX, wantX) || !closeEnough(gotY, wantY) || !closeEnough(gotZ, wantZ) {
			t.Fatalf("rotation mismatch: got (%v, %v, %v), want (%v, %v, %v)", gotX, gotY, gotZ, wantX, wantY, wantZ)
		}
	}
}

func TestYMPlayerSeekUsesPCMByteOffsets(t *testing.T) {
	player, err := NewYMPlayer(musicData, audioSampleRate, true)
	if err != nil {
		t.Fatal(err)
	}
	defer player.Close()

	want := int64(audioSampleRate * 4)
	got, err := player.Seek(want+3, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("Seek returned %d, want aligned byte offset %d", got, want)
	}
	if position := player.player.GetPos(); position < 990 || position > 1010 {
		t.Fatalf("StSound position is %d ms after seeking to one second", position)
	}
}

func TestYMPlayerReadDoesNotAllocate(t *testing.T) {
	player, err := NewYMPlayer(musicData, audioSampleRate, true)
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
		t.Fatalf("YMPlayer.Read allocated %.2f objects per call, want zero", allocations)
	}
}

func BenchmarkYMPlayerRead(b *testing.B) {
	player, err := NewYMPlayer(musicData, audioSampleRate, true)
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
