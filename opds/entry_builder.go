package opds

import (
	"time"

	"github.com/lann/builder"
)

type entryBuilder builder.Builder

func (e entryBuilder) Title(title string) entryBuilder {
	return builder.Set(e, "Title", title).(entryBuilder)
}

func (e entryBuilder) ID(id string) entryBuilder {
	return builder.Set(e, "ID", id).(entryBuilder)
}

func (e entryBuilder) AddLink(link Link) entryBuilder {
	return builder.Append(e, "Link", link).(entryBuilder)
}

func (e entryBuilder) Published(published time.Time) entryBuilder {
	return builder.Set(e, "Published", Time(published)).(entryBuilder)
}

func (e entryBuilder) Updated(updated time.Time) entryBuilder {
	return builder.Set(e, "Updated", Time(updated)).(entryBuilder)
}

func (e entryBuilder) Author(author *Person) entryBuilder {
	return builder.Set(e, "Author", author).(entryBuilder)
}

func (e entryBuilder) Summary(summary *Text) entryBuilder {
	return builder.Set(e, "Summary", summary).(entryBuilder)
}

func (e entryBuilder) Content(content *Text) entryBuilder {
	return builder.Set(e, "Content", content).(entryBuilder)
}

func (e entryBuilder) Series(series string) entryBuilder {
	return builder.Set(e, "DcSeries", series).(entryBuilder)
}

func (e entryBuilder) SeriesPosition(pos string) entryBuilder {
	return builder.Set(e, "DcSeriesPosition", pos).(entryBuilder)
}

func (e entryBuilder) Build() Entry {
	return builder.GetStruct(e).(Entry)
}

// EntryBuilder is a fluent immutable builder to build OPDS entries
var EntryBuilder = builder.Register(entryBuilder{}, Entry{}).(entryBuilder)
