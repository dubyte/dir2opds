package opds

import (
	"time"

	"github.com/lann/builder"
)

type AcquisitionFeed struct {
	*Feed
	Dc string `xml:"xmlns:dc,attr"`
}

type feedBuilder builder.Builder

func (f feedBuilder) Title(title string) feedBuilder {
	return builder.Set(f, "Title", title).(feedBuilder)
}

func (f feedBuilder) ID(id string) feedBuilder {
	return builder.Set(f, "ID", id).(feedBuilder)
}

func (f feedBuilder) AddLink(link Link) feedBuilder {
	return builder.Append(f, "Link", link).(feedBuilder)
}

func (f feedBuilder) Updated(updated time.Time) feedBuilder {
	return builder.Set(f, "Updated", Time(updated)).(feedBuilder)
}

func (f feedBuilder) Author(author Person) feedBuilder {
	return builder.Set(f, "Author", &author).(feedBuilder)
}

func (f feedBuilder) AddEntry(entry Entry) feedBuilder {
	return builder.Append(f, "Entry", &entry).(feedBuilder)
}

func (f feedBuilder) Build() Feed {
	return builder.GetStruct(f).(Feed)
}

// FeedBuilder is a fluent immutable builder to build OPDS Feeds
var FeedBuilder = builder.Register(feedBuilder{}, Feed{}).(feedBuilder)
