package scene

import (
	"fmt"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

// AnimationRecipe binds an authored animation name to a reusable DCK program.
// The production chooses its constants; the compiler owns field inheritance.
type AnimationRecipe func(name string) (geometry.PointAnimationSpec, error)

func actionVector(v *Vector3) *geometry.Vec3 {
	if v == nil {
		return nil
	}
	return &geometry.Vec3{X: v.X, Y: v.Y, Z: v.Z}
}

// CompileActions retains the difference between an omitted field and an
// explicit zero, and between inherited, cleared and replaced animations.
func CompileActions(authored []Action, recipe AnimationRecipe) ([]geometry.PointAction, error) {
	if recipe == nil {
		return nil, fmt.Errorf("vectorballs: missing animation recipe")
	}
	actions := make([]geometry.PointAction, len(authored))
	for i, source := range authored {
		if source.HasPos && source.Pos == nil || source.HasInitRot && source.InitRot == nil ||
			source.HasRot && source.Rot == nil || source.HasTr && source.Tr == nil {
			return nil, fmt.Errorf("vectorballs: action %d sets a missing vector", i)
		}
		action := geometry.PointAction{Frames: source.Frames,
			SetShape: source.HasShape, Shape: source.Shape,
			SetAnimations: source.HasAnimTypes, DropLast: source.DropLastAnimation}
		if source.HasPos {
			action.Position = actionVector(source.Pos)
		}
		if source.HasInitRot {
			action.InitialRotation = actionVector(source.InitRot)
		}
		if source.HasRot {
			action.RotationStep = actionVector(source.Rot)
		}
		if source.HasTr {
			action.TranslationStep = actionVector(source.Tr)
		}
		if source.HasText {
			index := source.TextIndex
			action.TextIndex = &index
		}
		for _, name := range source.AnimTypes {
			spec, err := recipe(name)
			if err != nil {
				return nil, fmt.Errorf("vectorballs: action %d: %w", i, err)
			}
			action.Animations = append(action.Animations, spec)
		}
		for _, name := range source.AppendAnimTypes {
			spec, err := recipe(name)
			if err != nil {
				return nil, fmt.Errorf("vectorballs: action %d: %w", i, err)
			}
			action.AppendAnimations = append(action.AppendAnimations, spec)
		}
		actions[i] = action
	}
	return actions, nil
}
