package entry

const (
	Include LineType = iota
	Full
	Suffix
)

type LineType int

type Line struct {
	Type    LineType
	Payload string
}

type Entry struct {
	Lines []*Line
}
