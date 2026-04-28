package tv

type GrowFlag uint8

const (
	GfGrowLoX GrowFlag = 1 << iota
	GfGrowLoY
	GfGrowHiX
	GfGrowHiY
	GfGrowAll GrowFlag = GfGrowLoX | GfGrowLoY | GfGrowHiX | GfGrowHiY
	GfGrowRel GrowFlag = 1 << 4
)
