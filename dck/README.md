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
go run ./dck/cmd/vectorballs -object sphere -segments 6
```

`-size` sets model dimensions and `-ball` selects a sprite from 0 to 120. Omitting `-object` runs the original choreography. From Go, use `SetObject(ObjectOptions{...})`, or construct a complete DCK `sprites.ProjectedObject` with an editable cube, pyramid, plane, flag, sphere or custom point source. Fill modes select edges, faces or interior points for cubes and pyramids; every object is drawn as balls. The sphere uses an even Fibonacci distribution with `4 × segments²` balls and accepts `segments` from 1 to 32. Each object owns its rotation, optional flag wave, projection and reusable point buffers, so several independently configured instances can share the same ball atlas.

When `-ball` is omitted, the sphere uses a 20-pixel cyan sprite through sixteen
segments and a 12-pixel one above that density. An explicit `-ball` overrides
the choice. From twelve segments, the preset masks its rear hemisphere and
scales sprites with perspective; DCK batches equal-image sprites in depth
order. Six segments (144 balls) is the recommended readable vectorball look.
The 32-segment/4,096-point mode is a dense experimental point surface.

Android keeps the authored choreography when launched normally. A cold launch
can preview any DCK object without changing that default:

```sh
adb shell am start -S -n com.olivierh59500.vectorballs/.MainActivity \
  --es dck_object sphere --ei dck_segments 6
```

Optional `dck_fill`, `dck_size` and `dck_ball` extras override the material
and geometry. An absent `dck_ball` selects the density-aware sprite; invalid
extras log a warning and leave the authored action sequence active. This
preview entry point is available only in the DCK Android binding.

Native object captures: `go test -tags dck_rendercheck ./dck`.
Thirty-six complete-frame captures taken before and after this migration match
pixel for pixel, including every object change and the water reflection. The
demo no longer converts the optional object's points between two local slice
types on each frame. The authored action script below remains a separate DCK
`geometry.PointSequence` and keeps its original timing.

The current DCK APK was installed on a Pixel 10a (Android 17/API 37). Its
authored default sequence produced 744 distinct presented-frame intervals:
p95 16.780 ms, maximum 17.072 ms, none above 20 ms. Process PSS was
198,956 KiB, including 101,156 KiB of graphics memory, and thermal status
remained 0. The Android window now sets `FLAG_KEEP_SCREEN_ON`; the Pixel stayed
awake during this unattended sample. Cube, pyramid, plane and flag retain
their exact desktop comparisons. The new sphere has five GPU captures,
including its reflection; none of these optional modes has yet been selected
in the Android application.
After this sphere addition, the updated APK's authored sequence had another
744 presented-frame intervals: p95 16.776 ms, maximum 16.961 ms and none above
20 ms. Android still reports `KEEP_SCREEN_ON` for the foreground window.
The DCK sphere preview was also measured directly on the Pixel. With the final
published APK, the recommended 144-ball mode had 744 intervals, p95 16.776 ms,
maximum 17.594 ms and none above 20 ms. The dense 4,096-point mode had a
372-interval final-APK spot check with p95 16.769 ms, maximum 16.903 ms and no
slow interval; the longer local-code run also had zero above 20 ms among 744
intervals. Before batching, 217 of 744 dense-mode intervals exceeded 20 ms,
with p95 33.411 ms. The dense output looks like a point surface, so the
144-ball mode is the visual default. These short traces do not measure long-run
thermal behavior or battery use. Android still reports `KEEP_SCREEN_ON` and
remained awake with a 30-second screen timeout.

The original choreography's shape-to-shape transitions now use
`geometry.PointMorph` through the shared sequence. A new morph begins at the
current XYZ pose, retains each ball's artwork index, and advances with the
original per-frame rounding. The sine grid, helicopter rotors, Y orbit and
bouncing position use the same configurable DCK point-scene family.
The per-frame XYZ rotation matrix now comes from
`geometry.RotateXYZScaled`; it preserves the original coefficient order while
removing the trigonometric matrix builder from this production source.

`geometry.PointSequence` now owns the full stage clock and the ordered effects.
The production's shape table and action data live in the pure Go `dck/scene`
package; `go test ./dck/scene` checks a complete action cycle without a GPU.
The DCK game keeps artwork, audio, projection, reflection and layer placement.
The independent state comparison in that test now checks two complete cycles
at every logical tick, including all model coordinates and ball indices.
`go test -bench 'Benchmark(Shared|Former)PointSequence$' -benchmem ./dck/scene`
compares controller CPU time and allocations without rendering.

See the [DCK effect configuration guide](../../../lib/democonstructionkit/docs/EFFECT_OPTIONS.md) for the shared API and examples.
