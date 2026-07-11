package opds

import (
	"github.com/lann/builder"
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

func (l linkBuilder) FacetGroup(group string) linkBuilder {
	return builder.Set(l, "FacetGroup", group).(linkBuilder)
}

func (l linkBuilder) ActiveFacet(active string) linkBuilder {
	return builder.Set(l, "ActiveFacet", active).(linkBuilder)
}

func (l linkBuilder) Build() Link {
	return builder.GetStruct(l).(Link)
}

// LinkBuilder is a fluent immutable builder to build OPDS Links
var LinkBuilder = builder.Register(linkBuilder{}, Link{}).(linkBuilder)
