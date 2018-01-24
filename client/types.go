package client

type CharSet struct {
	value string
}

var (
	Utf_8 = CharSet{"UTF-8"}
	Iso_8859_1 = CharSet{"UTF-8"}
)

func (charSet CharSet) String() string {
	switch charSet {
	case Utf_8: return "UTF-8"
	case Iso_8859_1: return "ISO-8859-1"
	default: return charSet.value
	}
}

