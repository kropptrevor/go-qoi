package qoi

const (
	ChannelsRGB      uint8 = 3
	ChannelsRGBA     uint8 = 4
	ColorSpaceSRGB   uint8 = 0
	ColorSpaceLinear uint8 = 1
)

const (
	TagRGB   byte = 0b11111110
	TagRGBA  byte = 0b11111111
	TagIndex byte = 0b00_000000
	TagDiff  byte = 0b01_000000
	TagLuma  byte = 0b10_000000
	TagRun   byte = 0b11_000000
)
