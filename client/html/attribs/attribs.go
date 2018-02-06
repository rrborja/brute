package attribs

import "fmt"

type AttribValue interface {
	String() string
}

type Class string
type Id string

type TagAttr interface {
	Name() string
	Value() interface{}
	String() string
}

type Attr struct {
	name string
	value interface{}
}

func NewAttr(name string, value interface{}) *Attr {
	return &Attr{name, value}
}

func (attr *Attr) Name() string {
	return attr.name
}

func (attr *Attr) Value() interface{} {
	return attr.value
}

func (attr Attr) String() string {
	switch v := attr.value.(type) {
	case bool: if v { return fmt.Sprintf(`%s`, attr.Name()) }
	default: return fmt.Sprintf(`%s="%s"`, attr.Name(), attr.Value())
	}
	return ""
}

type SelfAttr struct {
	name string
}

func NewSelfAttr(name string) *SelfAttr {
	return &SelfAttr{name}
}

func (selfAttr *SelfAttr) Name() string {
	return selfAttr.name
}

func (selfAttr *SelfAttr) Value() interface{} {
	return nil
}

func (selfAttr SelfAttr) String() string {
	return selfAttr.Name()
}