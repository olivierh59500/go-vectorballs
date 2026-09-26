package scene

import (
	"math"
)

const FrameRateConversion = 2.4

// Action represents a timeline action
type Action struct {
	Shape             string
	Pos               *Vector3
	InitRot           *Vector3
	Rot               *Vector3
	Tr                *Vector3
	AnimTypes         []string
	Frames            int
	TextIndex         int
	AppendAnimTypes   []string
	DropLastAnimation int
	HasText           bool
	HasShape          bool
	HasPos            bool
	HasInitRot        bool
	HasRot            bool
	HasTr             bool
	HasAnimTypes      bool // true if AnimTypes was explicitly set (even if empty)
}

// AuthoredActions returns the original stage data without animation logic.
func AuthoredActions() []Action {
	toRad := math.Pi / 180
	toRadSpeed := toRad / FrameRateConversion

	return []Action{
		// SQUARE
		{Shape: "square", Pos: &Vector3{X: 0, Y: 320, Z: 850}, InitRot: &Vector3{X: 30 * toRad, Y: 30 * toRad, Z: 30 * toRad},
			Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, AnimTypes: []string{}, Frames: 45, TextIndex: 0,
			HasShape: true, HasPos: true, HasInitRot: true, HasRot: true, HasTr: true, HasAnimTypes: true, HasText: true},
		{Pos: &Vector3{X: 0, Y: 0, Z: 850}, Tr: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 45, HasPos: true, HasTr: true},
		{Frames: 30, TextIndex: 1, HasText: true},

		// RIPP_IN + RIPPLE
		{AnimTypes: []string{"Sinus2D", "YRotate"}, Frames: 60, HasAnimTypes: true},
		{Frames: 60, TextIndex: 2, HasText: true},
		{Frames: 60, TextIndex: 3, HasText: true},
		{Frames: 120},

		// MERGE to sphere
		{AnimTypes: []string{"MorphingSphere"}, Frames: 45, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 4 * toRadSpeed}, HasRot: true, HasAnimTypes: true},
		{AnimTypes: []string{}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 80, TextIndex: 4, HasRot: true, HasAnimTypes: true, HasText: true},
		{Frames: 40, TextIndex: 5, HasText: true},
		{AnimTypes: []string{"YRotate"}, Frames: 240, HasAnimTypes: true},

		// TUBE
		{AnimTypes: []string{"MorphingTube"}, Frames: 45, TextIndex: 6, HasText: true, HasAnimTypes: true},
		{Frames: 60},
		{Shape: "tube", Frames: 60, TextIndex: 7, HasShape: true, HasText: true},
		{Frames: 80},
		{AnimTypes: []string{"MorphingSquare"}, Frames: 60, TextIndex: 8, HasText: true, HasAnimTypes: true},
		{AnimTypes: []string{}, Frames: 50, TextIndex: 9, HasText: true, HasAnimTypes: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Frames: 50, HasTr: true},

		// HELI
		{Shape: "heli", AnimTypes: []string{"Rotors"}, Pos: &Vector3{X: 0, Y: 320, Z: 850}, InitRot: &Vector3{X: 180 * toRad, Y: 0, Z: 0},
			Tr: &Vector3{X: 0, Y: -3, Z: 0}, Rot: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 45, TextIndex: 10,
			HasShape: true, HasPos: true, HasInitRot: true, HasTr: true, HasRot: true, HasAnimTypes: true, HasText: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 60, HasTr: true},
		{Rot: &Vector3{X: 0, Y: -3 * toRadSpeed, Z: 0}, Frames: 30, HasRot: true},
		{Frames: 60, TextIndex: 11, HasText: true},
		{Rot: &Vector3{X: toRadSpeed, Y: 0, Z: toRadSpeed}, Frames: 20, HasRot: true},
		{AppendAnimTypes: []string{"YRotate"}, Rot: &Vector3{X: 0, Y: math.Pi * toRadSpeed, Z: 0}, Frames: 240, HasRot: true},
		{Rot: &Vector3{X: math.Pi * toRadSpeed, Y: math.Pi * toRadSpeed, Z: 0}, Frames: 220, TextIndex: 12, HasRot: true, HasText: true},
		{DropLastAnimation: 1, Rot: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 20, HasRot: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Frames: 45, HasTr: true},

		// ANIMAL - IMPORTANT: anim:[] clears the Rotors animation from helicopter!
		{Shape: "animal", AnimTypes: []string{}, Pos: &Vector3{X: 0, Y: 320, Z: 850}, InitRot: &Vector3{X: 180 * toRad, Y: 90 * toRad, Z: 0},
			Tr: &Vector3{X: 0, Y: -3, Z: 0}, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 0}, Frames: 45, TextIndex: 13,
			HasShape: true, HasPos: true, HasInitRot: true, HasTr: true, HasRot: true, HasAnimTypes: true, HasText: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 0, Y: 3 * toRadSpeed, Z: 0}, Frames: 180, HasTr: true, HasRot: true},
		{Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 120, TextIndex: 14, HasRot: true, HasText: true},
		{Tr: &Vector3{X: 0, Y: -1.25, Z: 0}, Frames: 120, HasTr: true},

		// BUBS
		{Shape: "bubs", Pos: &Vector3{X: 0, Y: -320, Z: 850}, Rot: &Vector3{X: 0, Y: 5 * toRadSpeed, Z: 0}, Tr: &Vector3{X: 0, Y: 1.5, Z: 0}, Frames: 240, TextIndex: 15,
			HasShape: true, HasPos: true, HasRot: true, HasTr: true, HasText: true},

		// MAN
		{Shape: "man", Pos: &Vector3{X: 0, Y: 320, Z: 850}, Rot: &Vector3{X: 0, Y: 4 * toRadSpeed, Z: 0}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, Frames: 45, TextIndex: 16,
			HasShape: true, HasPos: true, HasRot: true, HasTr: true, HasText: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 2 * toRadSpeed, Y: 4 * toRadSpeed, Z: 2 * toRadSpeed}, Frames: 240, HasTr: true, HasRot: true},
		{Frames: 120, TextIndex: 17, HasText: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Frames: 45, HasTr: true},

		// WOMAN
		{Shape: "woman", Pos: &Vector3{X: 0, Y: 320, Z: 850}, Rot: &Vector3{X: -2 * toRadSpeed, Y: 4 * toRadSpeed, Z: -2 * toRadSpeed}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, Frames: 45,
			HasShape: true, HasPos: true, HasRot: true, HasTr: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 0, Y: 4 * toRadSpeed, Z: 2 * toRadSpeed}, Frames: 360, TextIndex: 18, HasTr: true, HasRot: true, HasText: true},
		{Tr: &Vector3{X: 0, Y: -3, Z: 0}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 4 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 45, HasTr: true, HasRot: true},

		// TARDI transitions
		{Shape: "tardi", Pos: &Vector3{X: 0, Y: 0, Z: 850}, InitRot: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 0, Y: 3 * toRadSpeed, Z: 0}, Tr: &Vector3{X: 0, Y: 0, Z: 0}, Frames: 20, TextIndex: 19,
			HasShape: true, HasPos: true, HasInitRot: true, HasRot: true, HasTr: true, HasText: true},
		{Frames: 10}, {Frames: 10}, {Frames: 10}, {Frames: 10}, {Frames: 10}, {Frames: 10},

		// TARDI
		{Shape: "tardi", Frames: 90, HasShape: true},
		{Frames: 90, TextIndex: 20, HasText: true},
		{Rot: &Vector3{X: -3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 0}, AnimTypes: []string{"YRotate"}, Frames: 120, TextIndex: 21, HasRot: true, HasAnimTypes: true, HasText: true},
		{Tr: &Vector3{X: 0, Y: -2, Z: 0}, Frames: 120, TextIndex: 22, HasTr: true, HasText: true},

		// SPIDER
		{Shape: "spider", Tr: &Vector3{X: 0, Y: 0, Z: 0}, Pos: &Vector3{X: 0, Y: -900, Z: 850}, InitRot: &Vector3{X: 0, Y: 180 * toRad, Z: 0},
			Rot: &Vector3{X: 0, Y: 3 * toRadSpeed, Z: 0}, AnimTypes: []string{"Bounce"}, Frames: 215, TextIndex: 23,
			HasShape: true, HasTr: true, HasPos: true, HasInitRot: true, HasRot: true, HasAnimTypes: true, HasText: true},
		{AnimTypes: []string{}, Frames: 75, TextIndex: 24, HasAnimTypes: true, HasText: true},
		{Frames: 75, TextIndex: 25, HasText: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Rot: &Vector3{X: 0, Y: 2 * toRadSpeed, Z: 0}, Frames: 45, TextIndex: 26, HasTr: true, HasRot: true, HasText: true},

		// CUBE
		{Shape: "cube", Pos: &Vector3{X: 0, Y: 320, Z: 850}, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 4 * toRadSpeed}, Tr: &Vector3{X: 0, Y: -3, Z: 0}, Frames: 45,
			HasShape: true, HasPos: true, HasRot: true, HasTr: true},
		{Tr: &Vector3{X: 0, Y: 0, Z: 0}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 240, TextIndex: 27, HasTr: true, HasRot: true, HasText: true},
		{AnimTypes: []string{"MorphingSpaceCube"}, Frames: 45, HasAnimTypes: true},
		{AnimTypes: []string{}, Rot: &Vector3{X: 3 * toRadSpeed, Y: 3 * toRadSpeed, Z: 3 * toRadSpeed}, Frames: 120, HasRot: true, HasAnimTypes: true},
		{Tr: &Vector3{X: 0, Y: 3, Z: 0}, Rot: &Vector3{X: 4 * toRadSpeed, Y: 4 * toRadSpeed, Z: 4 * toRadSpeed}, Frames: 45, HasTr: true, HasRot: true},
	}
}
