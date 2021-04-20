package overlays

import (
	"github.com/wieku/danser-go/app/dance/movers"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/shape"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"

	"math"
)

type MomentumOverlay struct {
	nFont  *font.Font
	shaper *shape.Renderer
	mover  *movers.MomentumMover
}

func NewMomentumOverlay(mover *movers.MomentumMover) *MomentumOverlay {
	overlay := new(MomentumOverlay)

	overlay.mover = mover

	overlay.nFont = font.GetFont("Exo 2 Bold")
	overlay.shaper = shape.NewRenderer()

	return overlay
}

func (overlay *MomentumOverlay) Update(float64) {}

func (overlay *MomentumOverlay) DrawBeforeObjects(_ *batch.QuadBatch, _ []color2.Color, _ float64) {}

func (overlay *MomentumOverlay) DrawNormal(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	m := overlay.mover
	ms := settings.Dance.Momentum
	area := float32(ms.RestrictArea * math.Pi / 180.0)
	angle := float32(ms.RestrictAngle * math.Pi / 180.0)

	batch.Flush()

	overlay.shaper.SetCamera(batch.Projection)

	overlay.shaper.Begin()

	if m.A2 != -999 {
		overlay.shaper.SetColor(1, 0, 0, 1)
		overlay.shaper.DrawLineV(m.P3, m.P3.Add(vector.NewVec2fRad(m.A2 + math.Pi, m.P3.Dst(m.P0))), 4)

		var pt, mt float32 // thickness
		if m.AP {
			pt = 6
			mt = 2
		} else {
			pt = 2
			mt = 6
		}

		aA := m.P0.AngleRV(m.P3)

		overlay.shaper.SetColor(1, 1, 1, 1) // middle
		overlay.shaper.DrawLineV(m.P3, m.P3.Add(vector.NewVec2fRad(aA, 2000)), 2)

		overlay.shaper.SetColor(1, 0, 1, 1)
		overlay.shaper.DrawLineV(m.P3, m.P3.Add(vector.NewVec2fRad(aA + area, 2000)), pt)

		overlay.shaper.SetColor(1, 0, 1, 1)
		overlay.shaper.DrawLineV(m.P3, m.P3.Add(vector.NewVec2fRad(aA - area, 2000)), mt)

		a := m.P3.AngleRV(m.P0)

		overlay.shaper.SetColor(1, 1, 0, 1)
		overlay.shaper.DrawLineV(m.P3, m.P3.Add(vector.NewVec2fRad(a - angle, 2000)), pt)

		overlay.shaper.SetColor(1, 1, 0, 1)
		overlay.shaper.DrawLineV(m.P3, m.P3.Add(vector.NewVec2fRad(a + angle, 2000)), mt)
	}

	// bézier path
	overlay.shaper.SetColor(0, 1, 0, 1)
	overlay.shaper.DrawLineV(m.P0, m.P1, 3)
	overlay.shaper.DrawLineV(m.P1, m.P2, 3)
	overlay.shaper.DrawLineV(m.P2, m.P3, 5)

	overlay.shaper.End()
}

func (overlay *MomentumOverlay) DrawHUD(batch *batch.QuadBatch, _ []color2.Color, _ float64) {
	batch.SetColor(0, 1, 0, 1)
	overlay.nFont.DrawMonospaced(batch, 20, 20, 40, "Mover Path")

	if overlay.mover.A2 != -999 {
		batch.SetColor(1, 0, 0, 1)
		overlay.nFont.DrawMonospaced(batch, 20, 60, 40, "Next Obj Angle")

		batch.SetColor(1, 0, 1, 1)
		overlay.nFont.DrawMonospaced(batch, 20, 100, 40, "RestrictArea")

		batch.SetColor(1, 1, 0, 1)
		overlay.nFont.DrawMonospaced(batch, 20, 140, 40, "RestrictAngle")
	}
}

func (overlay *MomentumOverlay) IsBroken(_ *graphics.Cursor) bool {
	return false
}

func (overlay *MomentumOverlay) DisableAudioSubmission(_ bool) {}
