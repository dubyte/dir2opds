package opds

import (
	"github.com/lann/builder"
)

type authorBuilder builder.Builder

func (a authorBuilder) Name(name string) authorBuilder {
	return builder.Set(a, "Name", name).(authorBuilder)
}

func (a authorBuilder) URI(uri string) authorBuilder {
	return builder.Set(a, "URI", uri).(authorBuilder)
}

func (a authorBuilder) Email(email string) authorBuilder {
	return builder.Set(a, "Email", email).(authorBuilder)
}

func (a authorBuilder) InnerXML(inner string) authorBuilder {
	return builder.Set(a, "InnerXML", inner).(authorBuilder)
}

func (a authorBuilder) Build() Person {
	return builder.GetStruct(a).(Person)
}

// AuthorBuilder is a fluent immutable builder to build OPDS Authors
var AuthorBuilder = builder.Register(authorBuilder{}, Person{}).(authorBuilder)
