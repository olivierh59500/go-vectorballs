# DCK version

This directory contains the construction-kit version of go-vectorballs. The original Go sources are preserved at their original paths (revision `22ccd09dc8c81a6ae1ff1b2c99fcbf68935d0b42`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/vectorballs` and this version with `go run ./dck/cmd/vectorballs` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Music is opened with `sound.Open`; DCK selects the decoder from the asset and
provides the configured stereo PCM format. The demo keeps its playback level and loop settings.

## Predefined vectorball objects

```sh
go run ./dck/cmd/vectorballs -object cube -fill edges
go run ./dck/cmd/vectorballs -object cube -fill solid -segments 4
go run ./dck/cmd/vectorballs -object pyramid -fill surface
go run ./dck/cmd/vectorballs -object plane
go run ./dck/cmd/vectorballs -object flag -segments 12
```

`-size` sets model dimensions and `-ball` selects a sprite from 0 to 120. Omitting `-object` runs the original choreography. From Go, use `SetObject(ObjectOptions{...})`, or use the independent DCK `sprites.Cube`, `Pyramid`, `Plane` and `Flag` directly with `sprites.Projector`. Fill modes select edges, faces or interior points; all are drawn as balls.

Native object captures: `go test -tags dck_rendercheck ./dck`.

The original choreography's shape-to-shape transitions now use
`geometry.PointMorph` through the shared sequence. A new morph begins at the
current XYZ pose, retains each ball's artwork index, and advances with the
original per-frame rounding. The sine grid, helicopter rotors, Y orbit and
bouncing position use the same configurable DCK point-scene family.

`geometry.PointSequence` now owns the full stage clock and the ordered effects.
The production's shape table and action data live in the pure Go `dck/scene`
package; `go test ./dck/scene` checks a complete action cycle without a GPU.
The DCK game keeps artwork, audio, projection, reflection and layer placement.
The independent state comparison in that test now checks two complete cycles
at every logical tick, including all model coordinates and ball indices.
`go test -bench 'Benchmark(Shared|Former)PointSequence$' -benchmem ./dck/scene`
compares controller CPU time and allocations without rendering.

See the [DCK effect configuration guide](../../../lib/democonstructionkit/docs/EFFECT_OPTIONS.md) for the shared API and examples.
