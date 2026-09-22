# DCK version

This directory contains the construction-kit version of go-vectorballs. The original Go sources are preserved at their original paths (revision `22ccd09dc8c81a6ae1ff1b2c99fcbf68935d0b42`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/vectorballs` and this version with `go run ./dck/cmd/vectorballs` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Go downloads the dependencies automatically, including
`github.com/olivierh59500/ym-player v1.0.0` for YM playback.

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

See the [DCK effect configuration guide](../../../lib/democonstructionkit/docs/EFFECT_OPTIONS.md) for the shared API and examples.
