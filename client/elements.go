package client

import "github.com/rrborja/brute/client/attribs"

type Element struct {
	Tag
	Id_ *attribs.Id
	Class_ []attribs.Class
	Attributes_ []Attr
	Content interface{}
}

type AfterIdElement Element

func (element *Element) Id(value string) *AfterIdElement {
	id := attribs.Id(value)
	element.Id_ = &id
	e := AfterIdElement(*element)
	return &e
}

func (element Element) Class(value string) *AfterIdElement {
	e := AfterIdElement(element)
	return e.Class(value)
}

func (element Element) Value(value interface{}) string {
	e := AfterIdElement(element)
	return e.Value(value)
}

func (element *AfterIdElement) Class(value string) *AfterIdElement {
	class := attribs.Class(value)
	element.Class_ = append(element.Class_, class)
	return element
}

func (element *AfterIdElement) Attributes(attribs ...Attr) *AfterIdElement {
	element.Attributes_ = attribs
	return element
}

type ListElement Element

func (element ListElement) Value(items ...interface{}) string {
	e := AfterIdElement(element)
	return e.Value(func() {
		for _, item := range items {
			element := new(Element)
			element.Tag = Tag{Name: li}
			element.Value(item)
		}
	})
}

type RenderElement struct {
	*AfterIdElement
}

func (element *AfterIdElement) Value(value interface{}) string {
	element.Content = value
	return evaluate(Element(*element))
}

type SelfElement Element

func (element SelfElement) Value() string {
	result := renderBeginTag(element.Tag, element.Id_, element.Class_, element.Attributes_)
	root.content = string(append([]byte(root.content), result...))
	return result
}