package html

import (
	"github.com/rrborja/brute/client/html/attribs"
	"fmt"
)

type Element struct {
	Id_ *attribs.Id
	Class_ []attribs.Class
	Attributes_ []attribs.TagAttr
	Content interface{}

	Tag
	stack *RenderStackHolder
}

type AfterIdElement Element

func (element Element) Value(value interface{}) string {
	e := AfterIdElement(element)
	return e.Value(value)
}

func (element *Element) Id(name string) *AfterIdElement {
	id := attribs.Id(name)
	element.Id_ = &id
	e := AfterIdElement(*element)
	return &e
}

func (element *Element) Class(value string) *AfterIdElement {
	e := AfterIdElement(*element)
	return e.Class(value)
}

func (element *AfterIdElement) Class(value string) *AfterIdElement {
	class := attribs.Class(value)
	element.Class_ = append(element.Class_, class)
	return element
}

func (element *AfterIdElement) Attributes(attribs ...attribs.TagAttr) *AfterIdElement {
	element.Attributes_ = attribs
	return element
}

type ListElement Element

func (element *ListElement) Value(items ...interface{}) string {
	e := AfterIdElement(*element)
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
	return element.stack.evaluate(Element(*element))
}

type SelfElement Element

func (element SelfElement) Value() string {
	result := renderBeginTag(element.Tag, element.Id_, element.Class_, element.Attributes_)
	element.stack.root.content = string(append([]byte(element.stack.root.content), result...))
	element.stack.evaluate(Element(element))
	return result
}

type FormElement Element
type AfterIdFormElement FormElement
type AfterClassFormElement AfterIdFormElement
type AfterActionFormElement AfterClassFormElement
type AfterMethodFormElement AfterActionFormElement
type AfterAttributesFormElement AfterMethodFormElement
type FinalFormElement AfterActionFormElement

func (element *FormElement) Id(name string) *AfterIdFormElement {
	id := attribs.Id(name)
	element.Id_ = &id
	e := AfterIdFormElement(*element)
	return &e
}

func (element *AfterIdFormElement) Class(name ...string) *AfterClassFormElement {
	e := AfterClassFormElement(*element)
	return e.Class(name...)
}

func (element *AfterClassFormElement) Class(name ...string) *AfterClassFormElement {
	classes := make([]attribs.Class, len(name))
	for _, c := range classes {
		element.Class_ = append(element.Class_, c)
	}
	e := AfterClassFormElement(*element)
	return &e
}

func (element *AfterClassFormElement) Action(name string) *AfterActionFormElement {
	e := FormElement(*element)
	return e.Action(name)
}

func (element *FormElement) Action(path string) *AfterActionFormElement {
	element.Attributes_ = []attribs.TagAttr{
		attribs.NewAttr("action", path),
	}
	castedElement := AfterActionFormElement(*element)
	return &castedElement
}

func (element *AfterActionFormElement) Get() *AfterMethodFormElement {
	element.Attributes_ = append(element.Attributes_,
		attribs.NewAttr("method", "get"),
	)
	castedElement := AfterMethodFormElement(*element)
	return &castedElement
}

func (element *AfterActionFormElement) Post() *AfterMethodFormElement {
	element.Attributes_ = append(element.Attributes_,
		attribs.NewAttr("method", "post"),
	)
	castedElement := AfterMethodFormElement(*element)
	return &castedElement
}

func (element *AfterMethodFormElement) Attributes(attrib ...attribs.TagAttr) *AfterAttributesFormElement {
	element.Attributes_ = append(element.Attributes_, attrib...)
	castedElement := AfterAttributesFormElement(*element)
	return &castedElement
}

func (element *AfterAttributesFormElement) Value(value interface{}, submitElement ...*SubmitElement) string {
	e := AfterMethodFormElement(*element)
	return e.Value(value, submitElement...)
}

func (element *AfterMethodFormElement) Value(value interface{}, submitElements ...*SubmitElement) string {
	e := AfterIdElement(*element)
	return e.Value(func() {
		Value(value)
		if len(submitElements) == 1 {
			if element.Id_ != nil && len(*element.Id_) > 0 {
				buttonId := attribs.Id(fmt.Sprintf("%s-button", *element.Id_))
				submitElements[0].Id_ = &buttonId
			}
			element.stack.evaluate(Element(*submitElements[0]))
		}
		if len(submitElements) > 1 {
			for i, submitElement := range submitElements {
				if element.Id_ != nil && len(*element.Id_) > 0 {
					buttonId := attribs.Id(fmt.Sprintf("%s-button-%d", *element.Id_, i))
					submitElement.Id_ = &buttonId
				}
				element.stack.evaluate(Element(*submitElement))
			}
		}
	})
}

type SubmitElement Element