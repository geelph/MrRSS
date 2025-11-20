package opml

import (
	"MrRSS/internal/models"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"strings"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Title string `xml:"title"`
}

type Body struct {
	Outlines []*Outline `xml:"outline"`
}

type Outline struct {
	Text     string     `xml:"text,attr"`
	Title    string     `xml:"title,attr"`
	Type     string     `xml:"type,attr"`
	XMLURL   string     `xml:"xmlUrl,attr"`
	HTMLURL  string     `xml:"htmlUrl,attr"`
	Outlines []*Outline `xml:"outline"` // Nested outlines
}

func Parse(r io.Reader) ([]models.Feed, error) {
	// Read all content to handle BOM
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	log.Printf("OPML Parse: Read %d bytes", len(content))

	if len(content) == 0 {
		return nil, errors.New("file content is empty")
	}

	// Strip UTF-8 BOM if present
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))

	var doc OPML
	decoder := xml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&doc); err != nil {
		log.Printf("OPML Parse: Decode error: %v", err)
		return nil, err
	}

	var feeds []models.Feed
	var extract func([]*Outline, string)
	extract = func(outlines []*Outline, category string) {
		for _, o := range outlines {
			if o.XMLURL != "" {
				title := o.Title
				if title == "" {
					title = o.Text
				}
				feeds = append(feeds, models.Feed{
					Title:    title,
					URL:      o.XMLURL,
					Category: category,
				})
			}

			newCategory := category
			if o.XMLURL == "" && o.Text != "" {
				if newCategory != "" {
					newCategory += "/" + o.Text
				} else {
					newCategory = o.Text
				}
			}

			if len(o.Outlines) > 0 {
				extract(o.Outlines, newCategory)
			}
		}
	}
	extract(doc.Body.Outlines, "")
	return feeds, nil
}

func Generate(feeds []models.Feed) ([]byte, error) {
	doc := OPML{
		Version: "1.0",
		Head: Head{
			Title: "MrRSS Subscriptions",
		},
	}

	for _, f := range feeds {
		currentOutlines := &doc.Body.Outlines

		if f.Category != "" {
			parts := strings.Split(f.Category, "/")
			for _, part := range parts {
				var found *Outline
				for _, o := range *currentOutlines {
					if o.XMLURL == "" && o.Text == part {
						found = o
						break
					}
				}
				if found == nil {
					found = &Outline{
						Text:  part,
						Title: part,
					}
					*currentOutlines = append(*currentOutlines, found)
				}
				currentOutlines = &found.Outlines
			}
		}

		*currentOutlines = append(*currentOutlines, &Outline{
			Text:   f.Title,
			Title:  f.Title,
			Type:   "rss",
			XMLURL: f.URL,
		})
	}

	return xml.MarshalIndent(doc, "", "  ")
}
