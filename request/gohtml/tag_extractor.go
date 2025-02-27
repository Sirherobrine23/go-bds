package gohtml

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const TagName = "html" // Gohtml tag for retriever from struct

var (
	_ ContentExtractor = TextExtractor
	_ ContentExtractor = TagExtractor("").Extractor
)

// different types of content extractor
type ContentExtractor func(selection *goquery.Selection) (string, error)

// Return trimed text from selection
func TextExtractor(selection *goquery.Selection) (string, error) {
	return strings.TrimSpace(selection.Text()), nil
}

type TagExtractor string // Process selection from attribute

// Return trimed text from selection
func (c TagExtractor) Extractor(selection *goquery.Selection) (string, error) {
	return strings.TrimSpace(selection.AttrOr(string(c), "")), nil
}

// determine whether we want to go to the next DOM level or stay in the current one
// and use attribute to get the value
func getContentExtractor(sel *goquery.Selection, tagValue string) (*goquery.Selection, ContentExtractor) {
	switch {
	case tagValue == "__self": // Return current selection
		return sel, TextExtractor
	case strings.HasSuffix(tagValue, ",attr") || strings.HasSuffix(tagValue, ", attr"): // compatiblity with [gitlab.com/tanqhnguyen/gohtml]
		return sel, TagExtractor(strings.TrimSpace(tagValue[:strings.LastIndex(tagValue, ",")-1])).Extractor
	case strings.Contains(tagValue, " = ") || strings.Contains(tagValue, "= ") || strings.Contains(tagValue, " ="): // New method to get attribute
		sep := strings.LastIndex(tagValue, "=")
		selector, attr := tagValue[:sep], tagValue[sep+1:]
		return sel.Find(strings.TrimSpace(selector)), TagExtractor(strings.TrimSpace(attr)).Extractor
	default: // Process tag selector without get attribute
		return sel.Find(tagValue), TextExtractor
	}
}
