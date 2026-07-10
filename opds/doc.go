// Package opds provides fluent immutable builders for generating OPDS 1.1 (Open Publication Distribution System)
// feeds in XML format. It is used by dir2opds, a self-hosted ebook server and digital library solution,
// to create navigation and acquisition feeds for ebook readers and OPDS clients.
//
// OPDS is a standard for cataloging and distributing digital publications such as EPUB, PDF, MOBI,
// and AZW3 files. This package supports building Atom feeds with the required OPDS extensions,
// including links for pagination, search, cover images, and book acquisition.
//
// # Builders
//
// All builders follow an immutable pattern using github.com/lann/builder. Each setter returns a new
// builder instance, allowing method chaining. Always call Build() at the end of the chain.
//
// Available builders:
//   - FeedBuilder: creates atom.Feed structures for navigation or acquisition feeds
//   - EntryBuilder: creates atom.Entry structures for individual books or folders
//   - LinkBuilder: creates atom.Link structures with relations like start, subsection, acquisition, search
//   - AuthorBuilder: creates atom.Person structures for book authors
//   - TextBuilder: creates atom.Text structures for summaries and descriptions
//
// # Example
//
//	feed := opds.FeedBuilder.
//		ID("/catalog").
//		Title("My Digital Library").
//		Updated(time.Now()).
//		AddLink(opds.LinkBuilder.Rel("start").Href("/").Type("application/atom+xml;profile=opds-catalog;kind=navigation").Build()).
//		Build()
//
//	acFeed := &opds.AcquisitionFeed{
//		Feed: &feed,
//		Dc:   "http://purl.org/dc/terms/",
//		Opds: "http://opds-spec.org/2010/catalog",
//	}
//
// # Link Relations
//
// Common OPDS link relations used with LinkBuilder:
//   - "start": root catalog link
//   - "search": link to OpenSearch description document
//   - "subsection": link to a sub-catalog (navigation)
//   - "http://opds-spec.org/acquisition": direct download link for a book
//   - "http://opds-spec.org/image": cover image link
//   - "http://opds-spec.org/image/thumbnail": thumbnail image link
//   - "first", "previous", "next", "last": pagination links
//
// # Content Types
//
// Standard OPDS media types:
//   - Navigation feed: application/atom+xml;profile=opds-catalog;kind=navigation
//   - Acquisition feed: application/atom+xml;profile=opds-catalog;kind=acquisition
//   - OpenSearch description: application/opensearchdescription+xml
//
// For more information about OPDS, visit https://opds-spec.org.
// For the full dir2opds project, visit https://github.com/dubyte/dir2opds.
package opds
