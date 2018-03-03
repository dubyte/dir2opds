package main

import (
	"github.com/lann/builder"
	"golang.org/x/tools/blog/atom"
)

type linkBuilder builder.Builder

func (l linkBuilder) Rel(rel string) linkBuilder {
	return builder.Set(l, "Rel", rel).(linkBuilder)
}

func (l linkBuilder) Href(href string) linkBuilder {
	return builder.Set(l, "Href", href).(linkBuilder)
}

func (l linkBuilder) Type(typeName string) linkBuilder {
	return builder.Set(l, "Type", typeName).(linkBuilder)
}

func (l linkBuilder) HrefLang(lang string) linkBuilder {
	return builder.Set(l, "HrefLang", lang).(linkBuilder)
}

func (l linkBuilder) Title(title string) linkBuilder {
	return builder.Set(l, "Title", title).(linkBuilder)
}

func (l linkBuilder) Length(length uint) linkBuilder {
	return builder.Set(l, "Length", length).(linkBuilder)
}

func (l linkBuilder) Build() atom.Link {
	return builder.GetStruct(l).(atom.Link)
}

// LinkBuilder is a fluent immutable builder to build OPDS Links
var LinkBuilder = builder.Register(linkBuilder{}, atom.Link{}).(linkBuilder)
