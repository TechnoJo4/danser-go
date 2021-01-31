package settings

import (
	"math"
	color2 "github.com/wieku/danser-go/framework/math/color"
)

var Cursor = initCursor()

func initCursor() *cursor {
	return &cursor{
		TrailStyle:   1,
		Style23Speed: 0.18,
		Style4Shift:  0.5,
		Colors: &cursorColors{
			EnableRainbow: true,
			EnableComboColoring: false,
			RainbowSpeed:  8,
			BaseColor: &hsv{
				0,
				1.0,
				1.0},
			EnableCustomHueOffset: false,
			HueOffset:             0,
			FlashToTheBeat:        false,
			FlashAmplitude:        0,
			currentHues:           nil,
		},
		EnableCustomTagColorOffset:  true,
		TagColorOffset:              -36,
		EnableTrailGlow:             true,
		EnableCustomTrailGlowOffset: true,
		TrailGlowOffset:             -36.0,
		ScaleToCS:                   false,
		CursorSize:                  18,
		CursorExpand:                false,
		ScaleToTheBeat:              false,
		ShowCursorsOnBreaks:         true,
		BounceOnEdges:               false,
		TrailScale:                  1.0,
		TrailEndScale:               0.4,
		TrailDensity:                0.5,
		TrailMaxLength:              2000,
		TrailRemoveSpeed:            1,
		GlowEndScale:                0.4,
		InnerLengthMult:             0.9,
		AdditiveBlending:            true,
		CursorRipples:               true,
	}
}

type cursor struct {
	TrailStyle                  int
	Style23Speed                float64
	Style4Shift                 float64
	Colors                      *cursorColors
	EnableCustomTagColorOffset  bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used
	TagColorOffset              float64 //-36, offset of the next tag cursor
	EnableTrailGlow             bool    //true
	EnableCustomTrailGlowOffset bool    //true, if enabled, value set below will be used, if not, HueOffset of previous iteration will be used (or offset of 180° for single cursor)
	TrailGlowOffset             float64 //-36, offset of the cursor trail glow
	ScaleToCS                   bool    //false, if enabled, cursor will scale to beatmap CS value
	CursorSize                  float64 //18, cursor radius in osu!pixels
	CursorExpand                bool    //Should cursor be scaled to 1.3 when clicked
	ScaleToTheBeat              bool    //true, cursor size is changing with music peak amplitude
	ShowCursorsOnBreaks         bool    //true
	BounceOnEdges               bool    //false
	TrailScale                  float64 //0.4
	TrailEndScale               float64 //0.4
	TrailDensity                float64 //0.5 - 1/TrailDensity = distance between trail points
	TrailMaxLength              int64   //2000 - maximum width (in osu!pixels) of cursortrail
	TrailRemoveSpeed            float64 //1.0 - trail removal multiplier, 0.5 means half the speed
	GlowEndScale                float64 //0.4
	InnerLengthMult             float64 //0.9 - if glow is enabled, inner trail will be shortened to 0.9 * length
	AdditiveBlending            bool
	CursorRipples               bool
}

type cursorColors struct {
	EnableRainbow         bool    //true
	EnableComboColoring   bool    //false
	RainbowSpeed          float64 //8, degrees per second
	BaseColor             *hsv    //0..360, if EnableRainbow is disabled then this value will be used to calculate base color
	EnableCustomHueOffset bool    //false, false means that every iteration has an offset of i*360/n
	HueOffset             float64 //0, custom hue offset for mirror collages
	FlashToTheBeat        bool    //true, objects size is changing with music peak amplitude
	FlashAmplitude        float64 //50, hue offset for flashes
	currentHues           []float64
	offsets               []float64
}

func (cl *cursorColors) Init(cursors int) {
	var offset float64 
	if cl.EnableCustomHueOffset {
		offset = cl.HueOffset
	} else {
		offset = 360.0 / float64(cursors)
	}

	hues := make([]float64, cursors)
	offsets := make([]float64, cursors)
	for i := range hues {
		hues[i] = cl.BaseColor.Hue
		offsets[i] = offset * float64(i)
	}
	cl.currentHues = hues
	cl.offsets = offsets
}

func (cl *cursorColors) Update(delta float64) {
	if cl.EnableRainbow {
		for i, hue := range cl.currentHues {
			hue += cl.RainbowSpeed / 1000.0 * delta
			hue = math.Mod(hue, 360)
			if hue < 0.0 { hue += 360.0 }
			cl.currentHues[i] = hue
		}
	}
}

func (cl *cursorColors) UpdateHit(idx int, hue float64) {
	if cl.EnableComboColoring {
		cl.currentHues[idx] = hue
	}
}

func (cr *cursor) GetColors(divides, cursors int, beatScale, alpha float64) []color2.Color {
	cl := cr.Colors
	flashOffset := 0.0
	if cl.FlashToTheBeat {
		flashOffset = cl.FlashAmplitude * (beatScale - 1.0) / (Audio.BeatScale - 1)
	}

	s := float32(cl.BaseColor.Saturation)
	v := float32(cl.BaseColor.Value)
	if !cr.EnableCustomTagColorOffset {
		colors := make([]color2.Color, cursors)
		for i, hue := range cl.currentHues {
			hue = math.Mod(hue + flashOffset + cl.offsets[i], 360)
			if hue < 0.0 { hue += 360.0 }
			colors[i] = color2.NewHSVA(float32(hue), s, v, float32(alpha))
		}
		return colors
	} else {
		colors := make([]color2.Color, cursors*divides)
		for i := 0; i < divides; i++ {
			for j, hue := range cl.currentHues {
				hue = math.Mod(hue + flashOffset + cl.offsets[j] + float64(i) * cr.TagColorOffset, 360)
				if hue < 0.0 { hue += 360.0 }
				colors[i*cursors+j] = color2.NewHSVA(float32(hue), s, v, float32(alpha))
			}
		}
		return colors
	}
}
