# DCK version

This directory contains the construction-kit version of go-vectorballs. The original Go sources are preserved at their original paths (revision `22ccd09dc8c81a6ae1ff1b2c99fcbf68935d0b42`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/vectorballs` and this version with `go run ./dck/cmd/vectorballs` from the repository root.

The choreography and assets stay local; reusable rendering and effects live in `../../lib/democonstructionkit`. Second Reality retains its original ST3 music synchronization.
