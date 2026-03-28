package opds

import (
	"time"

	"github.com/lann/builder"
	"golang.org/x/tools/blog/atom"
)

type AcquisitionFeed struct {
	*atom.Feed
	Dc   string `xml:"xmlns:dc,attr"`
	Opds string `xml:"xmlns:opds,attr"`
}

type feedBuilder builder.Builder

func (f feedBuilder) Title(title string) feedBuilder {
	return builder.Set(f, "Title", title).(feedBuilder)
}

func (f feedBuilder) ID(id string) feedBuilder {
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

// FeedBuilder is a fluent immutable builder to build OPDS Feeds
var FeedBuilder = builder.Register(feedBuilder{}, atom.Feed{}).(feedBuilder)
