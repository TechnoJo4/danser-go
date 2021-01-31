package movers

import (
	"math"

	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
)

// https://github.com/TechnoJo4/osu/blob/master/osu.Game.Rulesets.Osu/Replays/Movers/MomentumMover.cs

type MomentumMover struct {
	*basicMover

	curve *curves.Bezier

	last      vector.Vector2f
	first     bool
	wasStream bool
}

func NewMomentumMover() MultiPointMover {
	return &MomentumMover{basicMover: &basicMover{}}
}

func (mover *MomentumMover) Reset(diff *difficulty.Difficulty, id int) {
	mover.basicMover.Reset(diff, id)

	mover.first = true
	mover.last = vector.NewVec2f(0, 0)
}

func same(mods difficulty.Modifier, o1 objects.IHitObject, o2 objects.IHitObject, skipStackAngles bool) bool {
	return o1.GetStackedStartPositionMod(mods) == o2.GetStackedStartPositionMod(mods) || (skipStackAngles && o1.GetStartPosition() == o2.GetStartPosition())
}

func anorm(a float32) float32 {
	pi2 := 2 * math32.Pi
	a = math32.Mod(a, pi2)
	if a < 0 {
		a += pi2
	}

	return a
}

func anorm2(a float32) float32 {
	a = anorm(a)
	if a > math32.Pi {
		a = -(2*math32.Pi - a)
	}

	return a
}

func (mover *MomentumMover) SetObjects(objs []objects.IHitObject) int {
	ms := settings.CursorDance.MoverSettings.Momentum[mover.id%len(settings.CursorDance.MoverSettings.Momentum)]

	i := 0

	start, end := objs[i], objs[i+1]

	hasNext := false
	var next objects.IHitObject
	if len(objs) > 2 {
		if _, ok := objs[i+2].(*objects.Circle); ok {
			hasNext = true
		}
		next = objs[i+2]
	}

	startPos := start.GetStackedEndPositionMod(mover.diff.Mods)
	endPos := end.GetStackedStartPositionMod(mover.diff.Mods)

	dst := startPos.Dst(endPos)
	mult := ms.DistanceMult
	multEnd := ms.DistanceMultEnd

	var a2 float32
	fromLong := false
	for i++; i < len(objs); i++ {
		o := objs[i]
		if s, ok := o.(objects.ILongObject); ok {
			a2 = s.GetStartAngleMod(mover.diff.Mods)
			fromLong = true
			break
		}
		if i == len(objs)-1 {
			a2 = mover.last.AngleRV(startPos)
			break
		}
		if !same(mover.diff.Mods, o, objs[i+1], ms.SkipStackAngles) {
			a2 = o.GetStackedStartPositionMod(mover.diff.Mods).AngleRV(objs[i+1].GetStackedStartPositionMod(mover.diff.Mods))
			o2 := objs[i+1]
			if s2, ok := o2.(objects.ILongObject); ok && ms.SliderPredict {
				pos := o.GetStackedStartPositionMod(mover.diff.Mods)
				pos2 := o2.GetStackedStartPositionMod(mover.diff.Mods)
				s2a := s2.GetStartAngleMod(mover.diff.Mods)
				dst2 := pos.Dst(pos2)
				pos2 = vector.NewVec2fRad(s2a, dst2 * float32(multEnd)).Add(pos2)

				a2 = pos.AngleRV(pos2)
			} else {
				a2 = o.GetStackedStartPositionMod(mover.diff.Mods).AngleRV(o2.GetStackedStartPositionMod(mover.diff.Mods))
			}
			break
		}
	}

	var sq1, sq2 float32
	if next != nil {
		nextPos := next.GetStackedStartPositionMod(mover.diff.Mods)
		sq1 = startPos.DstSq(endPos)
		sq2 = endPos.DstSq(nextPos)
	}


	// stream detection logic stolen from spline mover
	stream := false
	if hasNext && !fromLong && ms.StreamRestrict {
		min := float32(25.0)
		max := float32(10000.0)

		if sq1 >= min && sq1 <= max && mover.wasStream || (sq2 >= min && sq2 <= max) {
			stream = true
		}
	}

	mover.wasStream = stream

	var a1 float32
	if s, ok := start.(objects.ILongObject); ok {
		a1 = s.GetEndAngleMod(mover.diff.Mods)
	} else if mover.first {
		a1 = a2 + math.Pi
	} else {
		a1 = startPos.AngleRV(mover.last)
	}


	ac := a2 - endPos.AngleRV(startPos)

	area := float32(ms.RestrictArea * math.Pi / 180.0)
	sarea := float32(ms.StreamArea * math.Pi / 180.0)

	if sarea > 0 && stream && anorm(ac) < anorm((2*math32.Pi)-sarea) {
		a := startPos.AngleRV(endPos)

		sangle := float32(0.5 * math.Pi)
		if anorm(a1-a) > math32.Pi {
			a2 = a - sangle
		} else {
			a2 = a + sangle
		}

		mult = ms.StreamMult
		multEnd = ms.StreamMultEnd
	} else if !fromLong && area > 0 && math32.Abs(anorm2(ac)) < area {
		a := endPos.AngleRV(startPos)

		if (anorm(a2-a) < float32(ms.RestrictAngle * math.Pi / 180.0)) != ms.RestrictInvert {
			a2 = a + float32(ms.RestrictAngleAdd * math.Pi / 180.0)
		} else {
			a2 = a - float32(ms.RestrictAngleSub * math.Pi / 180.0)
		}
	} else if next != nil && !fromLong && ms.InterpolateAngles {
		r := sq1 / (sq1 + sq2)
		if !ms.InvertAngleInterpolation {
			r = sq2 / (sq1 + sq2)
		}

		a := startPos.AngleRV(endPos)
		a2 = a + r*anorm2(a2-a)
		mult = ms.DistanceMultOut
		multEnd = ms.DistanceMultOutEnd
	}

	startTime := start.GetEndTime()
	endTime := end.GetStartTime()
	duration := endTime - startTime

	if ms.DurationTrigger > 0 && duration >= ms.DurationTrigger {
		mult *= ms.DurationMult * (duration / ms.DurationTrigger)
	}

	_, es := end.(*objects.Slider)
	bounce := !es && same(mover.diff.Mods, end, start, ms.SkipStackAngles)
    if ms.EqualPosBounce > 0 && bounce {
        a1 = startPos.AngleRV(mover.last)
        a2 = a1 + math32.Pi
        dst = mover.last.Dst(startPos)
        mult = ms.EqualPosBounce
        multEnd = ms.EqualPosBounce
    }

	p1 := vector.NewVec2fRad(a1, dst * float32(mult)).Add(startPos)
	p2 := vector.NewVec2fRad(a2, dst * float32(multEnd)).Add(endPos)

	mover.curve = curves.NewBezierNA([]vector.Vector2f{startPos, p1, p2, endPos})

	if !bounce {
		mover.last = p2
	}

	mover.startTime = start.GetEndTime()
	mover.endTime = end.GetStartTime()
	mover.first = false

	return 2
}

func (mover *MomentumMover) Update(time float64) vector.Vector2f {
	t := bmath.ClampF64((time-mover.startTime)/(mover.endTime-mover.startTime), 0, 1)
	return mover.curve.PointAt(float32(t))
}
