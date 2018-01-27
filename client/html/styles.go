package html

import "fmt"

type BasicCssStyle struct {

}

type Css3Style struct {
	Animation
	Background
}

type Animation struct {
	Name string
	Duration
	TimingFunction Function
	Delay Duration
	IterationCount Number
	Direction
	*FillMode
	*PlayState
}

type Background struct {
	Attachment
	Color
	Clip Box
	Image *URL
	Origin Box
	Position
	Repeat Tiled
	Size
}

type Border struct {
	Width Length
	Style [4]Style
	Color [4]Color

	//Rest are individual properties
	BottomLeftRadius [2]Length
	BottomRightRadius [2]Length
	BottomStyle Style
	BottomWidth Thickness
	BorderImage
}

type BorderImage struct {
	Source *URL
	Slice [4]LengthOrFill
	ImageWidth [4]LengthOrNumberOrAuto
	Outset LengthOrNumber

}

type GlobalKeyword string
var (
	Initial GlobalKeyword = "initial"
	Inherit GlobalKeyword = "inherit"
)

type Duration struct {
	Value uint
	Unit DurationUnit
}

type DurationUnit string
var (
	Seconds DurationUnit = "s"
	Milliseconds DurationUnit = "ms"
)

type Length struct {
	Value int
	Unit MeasurementUnit
}

type MeasurementUnit string
var (
	Pixels MeasurementUnit = "px"
	Inches MeasurementUnit = "in"
	Percent MeasurementUnit = "%"
	Millimeters MeasurementUnit = "mm"
	Points MeasurementUnit = "pt"
	Pica MeasurementUnit = "pc"
	Em MeasurementUnit = "em"
	Ex MeasurementUnit = "ex"
	Rem MeasurementUnit = "rem"
)

type Number struct {
	Value uint
	Infinite bool
}

type Function func(...interface{}) string
var (
	Linear Function = func(n ...interface{}) string { return "linear" }
	Ease Function = func(n ...interface{}) string { return "ease" }
	EaseIn Function = func(n ...interface{}) string { return "ease-in" }
	EaseOut Function = func(n ...interface{}) string { return "ease-out" }
	EaseInOut Function = func(n ...interface{}) string { return "ease-in-out" }
	CubicBezier Function = func(n ...interface{}) string { return fmt.Sprintf("cubic-bezier(%g,%g,%g,%g)", n...) }
)

type Direction string
var (
	Normal Direction = "normal"
	Reverse Direction = "reverse"
	Alternate Direction = "alternate"
	AlternateReverse Direction = "alternate-reverse"
)

type FillMode string
var (
	Forwards FillMode = "forwards"
	Backwards FillMode = "backwards"
	Both FillMode = "both"
)

type PlayState string
var (
	Paused PlayState = "paused"
	Running PlayState = "running"
)

type Attachment string
var (
	Scroll Attachment = "scroll"
	Fixed Attachment = "fixed"
)

type Color struct {
	Name *string
	*Function
	Hexa *uint32
	Transparent bool
}
var (
	Rgb255 Function = func(n ...interface{}) string { return fmt.Sprintf("rgb(%d,%d,%d)", n...) }
	Rgb100 Function = func(n ...interface{}) string { return fmt.Sprintf("rgb(%f%%,%f%%,%f%%)", n...) }
	Rgba255 Function = func(n ...interface{}) string { return fmt.Sprintf("rgba(%d,%d,%d,%f)", n...) }
	Rgba100 Function = func(n ...interface{}) string { return fmt.Sprintf("rgba(%f%%,%f%%,%f%%,%f)", n...) }
	Hsl Function = func(n ...interface{}) string { return fmt.Sprintf("hsl(%d,%d%%,%d%%", n...) }
	Hsla Function = func(n ...interface{}) string { return fmt.Sprintf("hsl(%d,%d%%,%d%%,%f)", n...) }
)

type Box string
var (
	BorderBox string = "border-box"
	PaddingBox string = "padding-box"
	ContentBox string = "content-box"
)

type URL func(string) string
var (
	Url URL = func(url string) string { return fmt.Sprintf(`url("%s")`, url) }
)

type Position struct {
	Length
	Position *interface{}
}

type Alignment string
var (
	Left string = "left"
	Center string = "center"
	Right string = "right"
)

func (position Position) String() string {
	var second string
	if position.Position != nil {
		switch p := position.Position.(type) {
		case Alignment: second = " " + string(p)
		case Length: second = fmt.Sprintf(" %v%s", p.Value, p.Unit)
		}
	}
	return fmt.Sprintf("%v%s%s", position.Value, position.Unit, second)
}

type Tiled string
var (
	Repeat Tiled = "repeat"
	RepeatX Tiled = "repeat-x"
	RepeatY Tiled = "repeat-y"
	NoRepeat Tiled = "no-repeat"
)

type Modifier string
var (
	Auto Modifier = "auto"
	Cover Modifier = "cover"
	Contain Modifier = "contain"
)

type Size struct {
	*Length
	Modifier
}

type Style string
var (
	Hidden Style = "hidden"
	Dotted Style = "dotted"
	Dashed Style = "dashed"
	Solid Style = "solid"
	Double Style = "double"
	Groove Style = "groove"
	Ridge Style = "ridge"
	Inset Style = "inset"
	Outset Style = "outset"
)

type Thickness struct{
	ThicknessType
	Length
}

type ThicknessType string
var (
	Thin ThicknessType = "thin"
	Medium ThicknessType = "medium"
	Thick ThicknessType = "thick"
)

type LengthOrFill struct {
	Length [4]Length
	Fill *bool
}

type LengthOrNumberOrAuto struct {
	*Length
	*Number
	Auto *bool
}

type LengthOrNumber struct {
	*Length
	*Number
}
