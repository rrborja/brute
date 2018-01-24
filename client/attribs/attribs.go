package attribs

type Value interface {
	String() string
}

type Class string
type Id string