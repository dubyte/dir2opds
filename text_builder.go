package main

import (
	"github.com/lann/builder"
	"golang.org/x/tools/blog/atom"
)

type textBuilder builder.Builder

func (t textBuilder) Type(textType string) textBuilder {
	return builder.Set(t, "Type", textType).(textBuilder)
}

func (t textBuilder) Body(body string) textBuilder {
	return builder.Set(t, "Body", body).(textBuilder)
}

func (t textBuilder) Build() atom.Text {
	return builder.GetStruct(t).(atom.Text)
}

var TextBuilder = builder.Register(textBuilder{}, atom.Text{}).(textBuilder)
