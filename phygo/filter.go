package phygo

type Filter struct {
	CategoryBits int
	MaskBits     int
	GroupIndex   int
}

func NewFilter() Filter {
	return Filter{
		CategoryBits: 0x0001,
		MaskBits:     0xffff,
		GroupIndex:   0,
	}
}
