package opds

import (
	"github.com/lann/builder"
)

type textBuilder builder.Builder

func (t textBuilder) Type(textType string) textBuilder {
	return builder.Set(t, "Type", textType).(textBuilder)
}

func (t textBuilder) Body(body string) textBuilder {
	return builder.Set(t, "Body", body).(textBuilder)
}

func (t textBuilder) Build() Text {
	return builder.GetStruct(t).(Text)
}

// TextBuilder is a fluent immutable builder to build OPDS texts
var TextBuilder = builder.Register(textBuilder{}, Text{}).(textBuilder)
