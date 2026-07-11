package opds

import (
	"encoding/xml"
	"time"
)

// TimeStr is an RFC 3339-formatted time.
type TimeStr string

// Time formats t as an RFC 3339 string.
func Time(t time.Time) TimeStr {
	return TimeStr(t.Format("2006-01-02T15:04:05-07:00"))
}

// Feed is an Atom feed.
type Feed struct {
	XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string   `xml:"title"`
	ID      string   `xml:"id"`
	Link    []Link   `xml:"link"`
	Updated TimeStr  `xml:"updated"`
	Author  *Person  `xml:"author"`
	Entry   []*Entry `xml:"entry"`
	Opds    string   `xml:"xmlns:opds,attr,omitempty"`
}

// Entry is an Atom entry.
type Entry struct {
	Title     string  `xml:"title"`
	ID        string  `xml:"id"`
	Link      []Link  `xml:"link"`
	Published TimeStr `xml:"published"`
	Updated   TimeStr `xml:"updated"`
	Author    *Person `xml:"author"`
	Summary   *Text   `xml:"summary"`
	Content   *Text   `xml:"content"`
}

// Link is an Atom link with optional OPDS 1.2 facet attributes.
type Link struct {
	Rel         string `xml:"rel,attr,omitempty"`
	Href        string `xml:"href,attr"`
	Type        string `xml:"type,attr,omitempty"`
	HrefLang    string `xml:"hreflang,attr,omitempty"`
	Title       string `xml:"title,attr,omitempty"`
	Length      uint   `xml:"length,attr,omitempty"`
	FacetGroup  string `xml:"http://opds-spec.org/2010/catalog facetGroup,attr,omitempty"`
	ActiveFacet string `xml:"http://opds-spec.org/2010/catalog activeFacet,attr,omitempty"`
}

// Person is an Atom person (author or contributor).
type Person struct {
	Name string `xml:"name"`
	URI  string `xml:"uri,omitempty"`
	Email string `xml:"email,omitempty"`
}

// Text is an Atom text construct (summary or content).
type Text struct {
	Type string `xml:"type,attr,omitempty"`
	Body string `xml:",chardata"`
}
