// package service provides a http handler that reads the path in the request.url and returns
// an xml document that follows the OPDS 1.1 standard
// https://specs.opds.io/opds-1.1.html
package service

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dubyte/dir2opds/opds"
	"golang.org/x/tools/blog/atom"
)

func init() {
	_ = mime.AddExtensionType(".mobi", "application/x-mobipocket-ebook")
	_ = mime.AddExtensionType(".epub", "application/epub+zip")
	_ = mime.AddExtensionType(".cbz", "application/x-cbz")
	_ = mime.AddExtensionType(".cbr", "application/x-cbr")
	_ = mime.AddExtensionType(".fb2", "text/fb2+xml")
	_ = mime.AddExtensionType(".pdf", "application/pdf")
}

const (
	pathTypeFile = iota
	pathTypeDirOfDirs
	pathTypeDirOfFiles
)

const (
	defaultPageSize = 50
	maxPageSize     = 200
)

const (
	ignoreFile       = true
	includeFile      = false
	currentDirectory = "."
	parentDirectory  = ".."
	hiddenFilePrefix = "."
)

type OPDS struct {
	TrustedRoot      string
	HideCalibreFiles bool
	HideDotFiles     bool
	NoCache          bool
	EnableCache      bool
	SortBy           string
	ShowCovers       bool
	MimeMap          map[string]string
	EnableSearch     bool
	ExtractMetadata  bool
	EnableHTML       bool
	BaseURL          string
	PageSize         int
	NoPagination     bool
}

type Catalog struct {
	ID       string
	Title    string
	Type     int
	Entries  []CatalogEntry
	Cover    string
	Total    int
	Page     int
	PageSize int
	ModTime  time.Time
}

type CatalogEntry struct {
	Name      string
	Type      int
	ModTime   time.Time
	Size      int64
	Title     string
	Author    string
	CoverPath string
}

type IsDirer interface {
	IsDir() bool
}

func isFile(e IsDirer) bool {
	return !e.IsDir()
}

const navigationType = "application/atom+xml;profile=opds-catalog;kind=navigation"

var TimeNow = timeNowFunc()

// Scan inspects the directory and builds a Catalog model
func (s OPDS) pageSize() int {
	if s.PageSize <= 0 {
		return defaultPageSize
	}
	if s.PageSize > maxPageSize {
		return maxPageSize
	}
	return s.PageSize
}

func parsePage(pageStr string) int {
	if pageStr == "" {
		return 1
	}
	page := 1
	if n, err := strconv.Atoi(pageStr); err == nil && n > 0 {
		page = n
	}
	return page
}

func etag(urlPath string, modTime time.Time, page int) string {
	h := sha256.New()
	h.Write([]byte(urlPath))
	h.Write([]byte(modTime.UTC().Format(time.RFC3339Nano)))
	h.Write([]byte(strconv.Itoa(page)))
	return `"` + hex.EncodeToString(h.Sum(nil))[:16] + `"`
}

func (s OPDS) Scan(fPath string, urlPath string, page int) (*Catalog, error) {
	dirEntries, err := os.ReadDir(fPath)
	if err != nil {
		return nil, err
	}

	dirInfo, err := os.Stat(fPath)
	if err != nil {
		return nil, err
	}

	catalog := &Catalog{
		ID:      urlPath,
		Title:   "Catalog in " + urlPath,
		Type:    getPathType(fPath),
		ModTime: dirInfo.ModTime(),
	}

	for _, entry := range dirEntries {
		if fileShouldBeIgnored(entry.Name(), s.HideCalibreFiles, s.HideDotFiles) {
			continue
		}

		if s.ShowCovers && (entry.Name() == "cover.jpg" || entry.Name() == "folder.jpg") {
			catalog.Cover = filepath.Join(urlPath, entry.Name())
			continue
		}

		entryPath := filepath.Join(fPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			slog.Error("error getting info for entry", "error", err)
			continue
		}

		catalog.Entries = append(catalog.Entries, CatalogEntry{
			Name:    entry.Name(),
			Type:    getPathType(entryPath),
			ModTime: info.ModTime(),
			Size:    info.Size(),
		})

		if info.ModTime().After(catalog.ModTime) {
			catalog.ModTime = info.ModTime()
		}

		if s.ExtractMetadata && !entry.IsDir() {
			idx := len(catalog.Entries) - 1
			title, author, coverPath := extractMetadata(entryPath)
			if title != "" {
				catalog.Entries[idx].Title = title
			}
			if author != "" {
				catalog.Entries[idx].Author = author
			}
			if coverPath != "" {
				catalog.Entries[idx].CoverPath = coverPath
			}
		}
	}

	s.sortEntries(catalog.Entries)

	total := len(catalog.Entries)
	pageSize := s.pageSize()
	if page < 1 {
		page = 1
	}

	// When NoPagination is enabled, show all entries on a single page
	if s.NoPagination {
		pageSize = total
		if pageSize == 0 {
			pageSize = 1 // Avoid division by zero
		}
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	catalog.Total = total
	catalog.Page = page
	catalog.PageSize = pageSize
	catalog.Entries = catalog.Entries[start:end]

	return catalog, nil
}

func extractMetadata(path string) (string, string, string) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".epub":
		return extractEpubMetadata(path)
	case ".pdf":
		title, author := extractPdfMetadata(path)
		return title, author, ""
	}
	return "", "", ""
}

func extractEpubMetadata(path string) (string, string, string) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", "", ""
	}
	defer r.Close()

	var opfPath string
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			opfPath = f.Name
			break
		}
	}

	if opfPath == "" {
		return "", "", ""
	}

	f, err := r.Open(opfPath)
	if err != nil {
		return "", "", ""
	}
	defer f.Close()

	opfContent, err := io.ReadAll(f)
	if err != nil {
		return "", "", ""
	}

	var opf struct {
		Metadata struct {
			Title   string `xml:"title"`
			Creator string `xml:"creator"`
		} `xml:"metadata"`
		Manifest struct {
			Items []struct {
				ID        string `xml:"id,attr"`
				Href      string `xml:"href,attr"`
				MediaType string `xml:"media-type,attr"`
			} `xml:"item"`
		} `xml:"manifest"`
	}

	decoder := xml.NewDecoder(bytes.NewReader(opfContent))
	if err := decoder.Decode(&opf); err != nil {
		return "", "", ""
	}

	// If standard unmarshal fails to get values due to namespaces
	if opf.Metadata.Title == "" || opf.Metadata.Creator == "" {
		decoder = xml.NewDecoder(bytes.NewReader(opfContent))
		decoder.DefaultSpace = "http://purl.org/dc/elements/1.1/"
		var opf2 struct {
			Metadata struct {
				Title   string `xml:"title"`
				Creator string `xml:"creator"`
			} `xml:"metadata"`
		}
		_ = decoder.Decode(&opf2)
		if opf2.Metadata.Title != "" {
			opf.Metadata.Title = opf2.Metadata.Title
		}
		if opf2.Metadata.Creator != "" {
			opf.Metadata.Creator = opf2.Metadata.Creator
		}
	}

	// Find cover image in manifest
	coverPath := findEpubCover(r, opf.Manifest.Items, opfPath)

	return opf.Metadata.Title, opf.Metadata.Creator, coverPath
}

func findEpubCover(r *zip.ReadCloser, items []struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}, opfPath string) string {
	// Common cover image IDs and properties
	coverIDs := []string{"cover", "cover-image", "coverimage", "coverimage"}
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	// Get the base directory of the OPF file
	opfDir := filepath.Dir(opfPath)

	// First, look for items with cover-related IDs
	for _, item := range items {
		itemID := strings.ToLower(item.ID)
		for _, coverID := range coverIDs {
			if strings.Contains(itemID, coverID) {
				for _, ext := range imageExtensions {
					if strings.HasSuffix(strings.ToLower(item.Href), ext) {
						return filepath.Join(opfDir, item.Href)
					}
				}
			}
		}
	}

	// Second, look for common cover image filenames in the EPUB
	coverNames := []string{"cover.jpg", "cover.jpeg", "cover.png", "cover.gif", "cover.webp"}
	for _, f := range r.File {
		name := strings.ToLower(f.Name)
		for _, coverName := range coverNames {
			if strings.HasSuffix(name, coverName) {
				return f.Name
			}
		}
	}

	// Third, look in common image directories
	for _, f := range r.File {
		name := strings.ToLower(f.Name)
		if strings.Contains(name, "images/") || strings.Contains(name, "oebps/images/") {
			for _, ext := range imageExtensions {
				if strings.HasSuffix(name, ext) {
					// Prefer files with "cover" in the name
					if strings.Contains(name, "cover") {
						return f.Name
					}
				}
			}
		}
	}

	return ""
}

func extractPdfMetadata(path string) (string, string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var title, author string
	// Only scan first 4KB to keep it fast
	maxLines := 100
	for i := 0; i < maxLines && scanner.Scan(); i++ {
		line := scanner.Text()
		if title == "" && strings.Contains(line, "/Title") {
			title = parsePdfValue(line, "/Title")
		}
		if author == "" && strings.Contains(line, "/Author") {
			author = parsePdfValue(line, "/Author")
		}
		if title != "" && author != "" {
			break
		}
	}
	return title, author
}

func parsePdfValue(line, key string) string {
	idx := strings.Index(line, key)
	if idx == -1 {
		return ""
	}
	start := strings.Index(line[idx:], "(")
	if start == -1 {
		return ""
	}
	end := strings.Index(line[idx+start:], ")")
	if end == -1 {
		return ""
	}
	return line[idx+start+1 : idx+start+end]
}

func (s OPDS) sortEntries(entries []CatalogEntry) {
	switch s.SortBy {
	case "date":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].ModTime.After(entries[j].ModTime)
		})
	case "size":
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Size > entries[j].Size
		})
	default: // name
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Name < entries[j].Name
		})
	}
}

func isBrowser(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html")
}

// Handler serves the content of a book file or
// returns an Acquisition Feed when the entries are documents or
// returns a Navigation Feed when the entries are other folders
func (s OPDS) Handler(w http.ResponseWriter, req *http.Request) error {
	var err error
	urlPath, err := url.PathUnescape(req.URL.Path)
	if err != nil {
		slog.Error("error unescaping path", "urlPath", req.URL.Path, "error", err)
		return err
	}

	fPath := filepath.Join(s.TrustedRoot, urlPath)

	// verifyPath avoid the http transversal by checking the path is under DirRoot
	_, err = verifyPath(fPath, s.TrustedRoot)
	if err != nil {
		slog.Error("verify path error", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	if _, err := os.Stat(fPath); err != nil {
		slog.Error("file system stat error", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	pathType := getPathType(fPath)

	// it's a file just serve the file
	if pathType == pathTypeFile {
		http.ServeFile(w, req, fPath)
		return nil
	}

	if s.NoCache {
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Add("Expires", "0")
	}

	page := parsePage(req.URL.Query().Get("page"))
	catalog, err := s.Scan(fPath, urlPath, page)
	if err != nil {
		slog.Error("error scanning path", "error", err)
		return err
	}

	slog.Debug("request",
		"urlPath", urlPath,
		"page", catalog.Page,
		"pageSize", catalog.PageSize,
		"total", catalog.Total,
		"totalPages", (catalog.Total+catalog.PageSize-1)/catalog.PageSize,
	)

	if s.EnableCache {
		eTag := etag(urlPath, catalog.ModTime, page)
		lastModified := catalog.ModTime.UTC()

		w.Header().Set("ETag", eTag)
		w.Header().Set("Last-Modified", lastModified.Format(http.TimeFormat))

		if ifNoneMatch := req.Header.Get("If-None-Match"); ifNoneMatch != "" {
			if ifNoneMatch == eTag {
				w.WriteHeader(http.StatusNotModified)
				return nil
			}
		}

		if ifModifiedSince := req.Header.Get("If-Modified-Since"); ifModifiedSince != "" {
			if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
				if !lastModified.After(t) {
					w.WriteHeader(http.StatusNotModified)
					return nil
				}
			}
		}
	}

	if s.EnableHTML && isBrowser(req) {
		return s.renderHTML(w, req, catalog)
	}

	navFeed := s.makeFeed(catalog, req)

	var content []byte
	// it is an acquisition feed
	if catalog.Type == pathTypeDirOfFiles {
		acFeed := &opds.AcquisitionFeed{Feed: &navFeed, Dc: "http://purl.org/dc/terms/", Opds: "http://opds-spec.org/2010/catalog"}
		content, err = xml.MarshalIndent(acFeed, "  ", "    ")
		w.Header().Add("Content-Type", "application/atom+xml;profile=opds-catalog;kind=acquisition")
	} else { // it is a navigation feed
		content, err = xml.MarshalIndent(navFeed, "  ", "    ")
		w.Header().Add("Content-Type", "application/atom+xml;profile=opds-catalog;kind=navigation")
	}
	if err != nil {
		slog.Error("error marshaling feed", "error", err)
		return err
	}

	content = append([]byte(xml.Header), content...)
	http.ServeContent(w, req, "feed.xml", TimeNow(), bytes.NewReader(content))

	return nil
}

// SearchHandler performs a basic filename search
func (s OPDS) SearchHandler(w http.ResponseWriter, req *http.Request) error {
	query := req.URL.Query().Get("q")
	if query == "" {
		return s.Handler(w, req)
	}

	page := parsePage(req.URL.Query().Get("page"))
	pageSize := s.pageSize()

	catalog := &Catalog{
		ID:    "search:" + query,
		Title: "Search results for: " + query,
		Type:  pathTypeDirOfFiles,
	}

	err := filepath.Walk(s.TrustedRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fileShouldBeIgnored(info.Name(), s.HideCalibreFiles, s.HideDotFiles) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(query)) {
			relPath, _ := filepath.Rel(s.TrustedRoot, path)
			catalog.Entries = append(catalog.Entries, CatalogEntry{
				Name:    relPath,
				Type:    pathTypeFile,
				ModTime: info.ModTime(),
				Size:    info.Size(),
			})
		}
		return nil
	})

	if err != nil {
		return err
	}

	s.sortEntries(catalog.Entries)

	total := len(catalog.Entries)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	catalog.Total = total
	catalog.Page = page
	catalog.PageSize = pageSize
	catalog.Entries = catalog.Entries[start:end]

	if s.EnableHTML && isBrowser(req) {
		return s.renderHTML(w, req, catalog)
	}

	navFeed := s.makeFeed(catalog, req)
	acFeed := &opds.AcquisitionFeed{Feed: &navFeed, Dc: "http://purl.org/dc/terms/", Opds: "http://opds-spec.org/2010/catalog"}
	content, err := xml.MarshalIndent(acFeed, "  ", "    ")
	if err != nil {
		return err
	}

	w.Header().Add("Content-Type", "application/atom+xml;profile=opds-catalog;kind=acquisition")
	content = append([]byte(xml.Header), content...)
	http.ServeContent(w, req, "feed.xml", TimeNow(), bytes.NewReader(content))
	return nil
}

// OpenSearchHandler serves the OpenSearch description document
func (s OPDS) OpenSearchHandler(w http.ResponseWriter, req *http.Request) {
	searchURL := s.joinURL("/search?q={searchTerms}")
	xmlStr := `<?xml version="1.0" encoding="UTF-8"?>
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
  <ShortName>dir2opds</ShortName>
  <Description>Search books in dir2opds</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <OutputEncoding>UTF-8</OutputEncoding>
  <Url type="application/atom+xml;profile=opds-catalog;kind=acquisition" template="` + searchURL + `"/>
</OpenSearchDescription>`
	w.Header().Set("Content-Type", "application/opensearchdescription+xml")
	w.Write([]byte(xmlStr))
}

func (s OPDS) joinURL(p string) string {
	if s.BaseURL == "" {
		return p
	}
	return strings.TrimSuffix(s.BaseURL, "/") + "/" + strings.TrimPrefix(p, "/")
}

// CoverHandler extracts and serves cover images from EPUB files
func (s OPDS) CoverHandler(w http.ResponseWriter, req *http.Request) error {
	filePath := req.URL.Query().Get("file")
	if filePath == "" {
		return fmt.Errorf("missing file parameter")
	}

	urlPath, err := url.PathUnescape(filePath)
	if err != nil {
		slog.Error("error unescaping cover path", "filePath", filePath, "error", err)
		return err
	}

	fPath := filepath.Join(s.TrustedRoot, urlPath)

	// verifyPath avoid the http transversal by checking the path is under TrustedRoot
	_, err = verifyPath(fPath, s.TrustedRoot)
	if err != nil {
		slog.Error("verify path error for cover", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	if _, err := os.Stat(fPath); err != nil {
		slog.Error("file stat error for cover", "error", err)
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	coverData, contentType, err := extractEpubCover(fPath)
	if err != nil {
		slog.Error("error extracting cover", "path", fPath, "error", err)
		return err
	}

	if coverData == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "max-age=86400")
	http.ServeContent(w, req, "cover", TimeNow(), bytes.NewReader(coverData))
	return nil
}

// extractEpubCover extracts the cover image from an EPUB file
func extractEpubCover(epubPath string) ([]byte, string, error) {
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, "", fmt.Errorf("opening epub: %w", err)
	}
	defer r.Close()

	var opfPath string
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			opfPath = f.Name
			break
		}
	}

	if opfPath == "" {
		return nil, "", fmt.Errorf("no OPF file found")
	}

	f, err := r.Open(opfPath)
	if err != nil {
		return nil, "", fmt.Errorf("opening OPF: %w", err)
	}
	defer f.Close()

	opfContent, err := io.ReadAll(f)
	if err != nil {
		return nil, "", fmt.Errorf("reading OPF: %w", err)
	}

	var opf struct {
		Manifest struct {
			Items []struct {
				ID        string `xml:"id,attr"`
				Href      string `xml:"href,attr"`
				MediaType string `xml:"media-type,attr"`
			} `xml:"item"`
		} `xml:"manifest"`
	}

	decoder := xml.NewDecoder(bytes.NewReader(opfContent))
	if err := decoder.Decode(&opf); err != nil {
		return nil, "", fmt.Errorf("parsing OPF: %w", err)
	}

	coverPath := findEpubCover(r, opf.Manifest.Items, opfPath)
	if coverPath == "" {
		return nil, "", nil
	}

	coverFile, err := r.Open(coverPath)
	if err != nil {
		return nil, "", fmt.Errorf("opening cover: %w", err)
	}
	defer coverFile.Close()

	coverData, err := io.ReadAll(coverFile)
	if err != nil {
		return nil, "", fmt.Errorf("reading cover: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(coverPath))
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return coverData, contentType, nil
}

func (s OPDS) makeFeed(catalog *Catalog, req *http.Request) atom.Feed {
	feedBuilder := opds.FeedBuilder.
		ID(catalog.ID).
		Title(catalog.Title).
		Updated(TimeNow()).
		AddLink(opds.LinkBuilder.Rel("start").Href(s.joinURL("/")).Type(navigationType).Build())

	if s.EnableSearch {
		feedBuilder = feedBuilder.AddLink(opds.LinkBuilder.
			Rel("search").
			Href(s.joinURL("/opensearch.xml")).
			Type("application/opensearchdescription+xml").
			Build())
	}

	if catalog.Cover != "" {
		coverHref := s.joinURL((&url.URL{Path: catalog.Cover}).String())
		feedBuilder = feedBuilder.AddLink(opds.LinkBuilder.
			Rel("http://opds-spec.org/image").
			Href(coverHref).
			Type(mime.TypeByExtension(filepath.Ext(catalog.Cover))).
			Build())
		feedBuilder = feedBuilder.AddLink(opds.LinkBuilder.
			Rel("http://opds-spec.org/image/thumbnail").
			Href(coverHref).
			Type(mime.TypeByExtension(filepath.Ext(catalog.Cover))).
			Build())
	}

	if !s.NoPagination && catalog.Total > catalog.PageSize {
		totalPages := (catalog.Total + catalog.PageSize - 1) / catalog.PageSize
		basePath := req.URL.Path
		query := req.URL.Query()

		feedType := "application/atom+xml;profile=opds-catalog;kind=navigation"
		if catalog.Type == pathTypeDirOfFiles {
			feedType = "application/atom+xml;profile=opds-catalog;kind=acquisition"
		}

		if catalog.Page > 1 {
			feedBuilder = feedBuilder.AddLink(opds.LinkBuilder.
				Rel("first").
				Href(s.joinURL(buildPageURL(basePath, query, 1))).
				Type(feedType).
				Build())
			feedBuilder = feedBuilder.AddLink(opds.LinkBuilder.
				Rel("previous").
				Href(s.joinURL(buildPageURL(basePath, query, catalog.Page-1))).
				Type(feedType).
				Build())
		}
		if catalog.Page < totalPages {
			feedBuilder = feedBuilder.AddLink(opds.LinkBuilder.
				Rel("next").
				Href(s.joinURL(buildPageURL(basePath, query, catalog.Page+1))).
				Type(feedType).
				Build())
			feedBuilder = feedBuilder.AddLink(opds.LinkBuilder.
				Rel("last").
				Href(s.joinURL(buildPageURL(basePath, query, totalPages))).
				Type(feedType).
				Build())
		}
	}

	for _, entry := range catalog.Entries {
		title := entry.Name
		if entry.Title != "" {
			title = entry.Title
		}

		var entryPath string
		if strings.HasPrefix(catalog.ID, "search:") {
			entryPath = "/" + entry.Name
		} else {
			entryPath = path.Join(req.URL.Path, entry.Name)
		}

		href := s.joinURL((&url.URL{Path: entryPath}).String())

		entryBuilder := opds.EntryBuilder.
			ID(req.URL.Path + entry.Name).
			Title(title).
			AddLink(opds.LinkBuilder.
				Rel(getRel(entry.Name, entry.Type)).
				Title(entry.Name).
				Href(href).
				Type(s.getType(entry.Name, entry.Type)).
				Build())

		if entry.Author != "" {
			entryBuilder = entryBuilder.Author(&atom.Person{Name: entry.Author})
		}

		if s.ExtractMetadata && entry.CoverPath != "" && entry.Type == pathTypeFile {
			coverURL := s.joinURL("/cover?file=" + url.QueryEscape(entryPath))
			ext := strings.ToLower(filepath.Ext(entry.CoverPath))
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "image/jpeg"
			}
			entryBuilder = entryBuilder.AddLink(opds.LinkBuilder.
				Rel("http://opds-spec.org/image").
				Href(coverURL).
				Type(contentType).
				Build())
			entryBuilder = entryBuilder.AddLink(opds.LinkBuilder.
				Rel("http://opds-spec.org/image/thumbnail").
				Href(coverURL).
				Type(contentType).
				Build())
		}

		feedBuilder = feedBuilder.AddEntry(entryBuilder.Build())
	}
	return feedBuilder.Build()
}

func buildPageURL(basePath string, query url.Values, page int) string {
	query.Set("page", strconv.Itoa(page))
	return basePath + "?" + query.Encode()
}

func fileShouldBeIgnored(filename string, hideCalibreFiles, hideDotFiles bool) bool {
	// not ignore those directories
	if filename == currentDirectory || filename == parentDirectory {
		return includeFile
	}

	if hideDotFiles && strings.HasPrefix(filename, hiddenFilePrefix) {
		return ignoreFile
	}

	if hideCalibreFiles &&
		(strings.Contains(filename, ".opf") ||
			strings.Contains(filename, "cover.") ||
			strings.Contains(filename, "metadata.db") ||
			strings.Contains(filename, "metadata_db_prefs_backup.json") ||
			strings.Contains(filename, ".caltrash") ||
			strings.Contains(filename, ".calnotes")) {
		return ignoreFile
	}

	return false
}

func getRel(name string, pathType int) string {
	if pathType == pathTypeDirOfFiles || pathType == pathTypeDirOfDirs {
		return "subsection"
	}

	ext := filepath.Ext(name)
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
		return "http://opds-spec.org/image/thumbnail"
	}

	// mobi, epub, etc
	return "http://opds-spec.org/acquisition"
}

func (s OPDS) getType(name string, pathType int) string {
	switch pathType {
	case pathTypeFile:
		ext := filepath.Ext(name)
		if s.MimeMap != nil {
			if mType, ok := s.MimeMap[ext]; ok {
				return mType
			}
		}
		return mime.TypeByExtension(ext)
	case pathTypeDirOfFiles:
		return "application/atom+xml;profile=opds-catalog;kind=acquisition"
	case pathTypeDirOfDirs:
		return "application/atom+xml;profile=opds-catalog;kind=navigation"
	default:
		return mime.TypeByExtension("xml")
	}
}

func getPathType(dirpath string) int {
	fi, err := os.Stat(dirpath)
	if err != nil {
		slog.Error("getPathType os.Stat error", "error", err)
		return pathTypeFile
	}

	if isFile(fi) {
		return pathTypeFile
	}

	dirEntries, err := os.ReadDir(dirpath)
	if err != nil {
		slog.Error("getPathType: readDir error", "error", err)
	}

	for _, entry := range dirEntries {
		if isFile(entry) {
			return pathTypeDirOfFiles
		}
	}
	// Directory of directories
	return pathTypeDirOfDirs
}

func timeNowFunc() func() time.Time {
	t := time.Now()
	return func() time.Time { return t }
}

// verifyPath uses trustedRoot to avoid http path traversal
// from https://www.stackhawk.com/blog/golang-path-traversal-guide-examples-and-prevention/
func verifyPath(path, trustedRoot string) (string, error) {
	// clean is already used upstream but leaving this
	// to keep the functionality of the function as close as possible to the blog.
	c := filepath.Clean(path)

	// get the canonical path
	r, err := filepath.EvalSymlinks(c)
	if err != nil {
		slog.Error("verifyPath error", "error", err)
		return c, errors.New("unsafe or invalid path specified")
	}

	if !inTrustedRoot(r, trustedRoot) {
		return r, errors.New("unsafe or invalid path specified")
	}

	return r, nil
}

func inTrustedRoot(path string, trustedRoot string) bool {
	path = filepath.Clean(path)
	trustedRoot = filepath.Clean(trustedRoot)
	if path == trustedRoot {
		return true
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(path, trustedRoot+sep)
}
