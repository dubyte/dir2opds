# opds - OPDS/Atom XML Builders

Immutable builders for OPDS 1.1 feed generation using `github.com/lann/builder`.

## WHERE TO LOOK

| Task | File | Key Functions |
|------|------|----------------|
| Create feed | `feed_builder.go` | `FeedBuilder.ID()`, `.Title()`, `.Updated()`, `.AddLink()`, `.AddEntry()` |
| Create entry | `entry_builder.go` | `EntryBuilder.ID()`, `.Title()`, `.AddLink()`, `.Author()` |
| Create link | `link_builder.go` | `LinkBuilder.Rel()`, `.Href()`, `.Type()`, `.Title()` |
| Create author | `author_builder.go` | `AuthorBuilder.Name()`, `.URI()`, `.Email()` |
| Create text | `text_builder.go` | `TextBuilder.Type()`, `.Body()` |

## BUILDER PATTERN

All builders follow the same immutable pattern:

```go
// Create feed
feed := opds.FeedBuilder.
    ID("/path").
    Title("Catalog").
    Updated(time.Now()).
    AddLink(opds.LinkBuilder.Rel("start").Href("/").Type("...").Build()).
    AddEntry(entryBuilder.Build()).
    Build()

// Create entry
entry := opds.EntryBuilder.
    ID("/book.epub").
    Title("My Book").
    AddLink(opds.LinkBuilder.
        Rel("http://opds-spec.org/acquisition").
        Href("/book.epub").
        Type("application/epub+zip").
        Build()).
    Build()

// Create acquisition feed (with namespaces)
acFeed := &opds.AcquisitionFeed{
    Feed: &feed,
    Dc:   "http://purl.org/dc/terms/",
    Opds: "http://opds-spec.org/2010/catalog",
}
```

## KEY TYPES

```go
type feedBuilder builder.Builder      // Unexported, use FeedBuilder singleton
type entryBuilder builder.Builder     // Unexported, use EntryBuilder singleton
type linkBuilder builder.Builder      // Unexported, use LinkBuilder singleton
type authorBuilder builder.Builder    // Unexported, use AuthorBuilder singleton
type textBuilder builder.Builder     // Unexported, use TextBuilder singleton

type AcquisitionFeed struct {
    *atom.Feed
    Dc   string `xml:"xmlns:dc,attr"`
    Opds string `xml:"xmlns:opds,attr"`
}
```

## CONVENTIONS

### Link Relations
- `start` - Link to root catalog
- `search` - Link to OpenSearch description
- `subsection` - Link to sub-catalog (navigation)
- `http://opds-spec.org/acquisition` - Download link
- `http://opds-spec.org/image` - Cover image
- `http://opds-spec.org/image/thumbnail` - Thumbnail
- `first`, `previous`, `next`, `last` - Pagination

### Content Types
- Navigation feed: `application/atom+xml;profile=opds-catalog;kind=navigation`
- Acquisition feed: `application/atom+xml;profile=opds-catalog;kind=acquisition`
- OpenSearch: `application/opensearchdescription+xml`

### Singleton Pattern
Each builder exports a singleton instance:
```go
var FeedBuilder = builder.Register(feedBuilder{}, atom.Feed{}).(feedBuilder)
var EntryBuilder = builder.Register(entryBuilder{}, atom.Entry{}).(entryBuilder)
var LinkBuilder = builder.Register(linkBuilder{}, atom.Link{}).(linkBuilder)
```

## ANTI-PATTERNS

- DO NOT create builder instances directly—use singleton (`FeedBuilder`, `EntryBuilder`, etc.)
- DO NOT forget to call `.Build()` at the end of the chain
- DO NOT reuse builder instances—chains are immutable, always create new