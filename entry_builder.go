package main

import (
	"github.com/lann/builder"
	"golang.org/x/tools/blog/atom"
	"time"
)

type entryBuilder builder.Builder

func (e entryBuilder) Title(title string) entryBuilder {
	return builder.Set(e, "Title", title).(entryBuilder)
}

func (e entryBuilder) Id(id string) entryBuilder {
	return builder.Set(e, "ID", id).(entryBuilder)
}

func (e entryBuilder) AddLink(link atom.Link) entryBuilder {
	return builder.Append(e, "Link", link).(entryBuilder)
}

func (e entryBuilder) Published(published time.Time) entryBuilder {
	return builder.Set(e, "Published", atom.Time(published)).(entryBuilder)
}

func (e entryBuilder) Updated(updated time.Time) entryBuilder {
	return builder.Set(e, "Updated", atom.Time(updated)).(entryBuilder)
}

func (e entryBuilder) Author(author *atom.Person) entryBuilder {
	return builder.Set(e, "Author", author).(entryBuilder)
}

func (e entryBuilder) Summary(summary *atom.Text) entryBuilder {
	return builder.Set(e, "Summary", summary).(entryBuilder)
}

func (e entryBuilder) Content(content *atom.Text) entryBuilder {
	return builder.Set(e, "Content", content).(entryBuilder)
}

func (e entryBuilder) Build() atom.Entry {
	return builder.GetStruct(e).(atom.Entry)
}

var EntryBuilder = builder.Register(entryBuilder{}, atom.Entry{}).(entryBuilder)
