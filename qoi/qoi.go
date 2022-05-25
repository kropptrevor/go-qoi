package qoi

import "image/color"

type Channels uint8

const (
	ChannelsRGB      Channels = 3
	ChannelsRGBA     Channels = 4
	ColorSpaceSRGB   uint8    = 0
	ColorSpaceLinear uint8    = 1
)

const (
	TagRGB   byte = 0b11111110
	TagRGBA  byte = 0b11111111
	TagMask  byte = 0b11_000000
	TagIndex byte = 0b00_000000
	TagDiff  byte = 0b01_000000
	TagLuma  byte = 0b10_000000
	TagRun   byte = 0b11_000000
)

type rgba color.NRGBA

func (color rgba) index() int {
	return int((color.R*3 + color.G*5 + color.B*7 + color.A*11) % 64)
}
