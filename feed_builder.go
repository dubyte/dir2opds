package main

import (
	"github.com/lann/builder"
	"golang.org/x/tools/blog/atom"
	"time"
)

type feedBuilder builder.Builder

func (f feedBuilder) Title(title string) feedBuilder {
	return builder.Set(f, "Title", title).(feedBuilder)
}

func (f feedBuilder) Id(id string) feedBuilder {
	return builder.Set(f, "ID", id).(feedBuilder)
}

func (f feedBuilder) AddLink(link atom.Link) feedBuilder {
	return builder.Append(f, "Link", link).(feedBuilder)
}

func (f feedBuilder) Updated(updated time.Time) feedBuilder {
	return builder.Set(f, "Updated", atom.Time(updated)).(feedBuilder)
}

func (f feedBuilder) Author(author atom.Person) feedBuilder {
	return builder.Set(f, "Author", &author).(feedBuilder)
}

func (f feedBuilder) AddEntry(entry atom.Entry) feedBuilder {
	return builder.Append(f, "Entry", &entry).(feedBuilder)
}

func (f feedBuilder) Build() atom.Feed {
	return builder.GetStruct(f).(atom.Feed)
}

var FeedBuilder = builder.Register(feedBuilder{}, atom.Feed{}).(feedBuilder)
