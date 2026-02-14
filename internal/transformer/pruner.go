package transformer

import (
	"strings"

	"golang.org/x/net/html"
)

type ImageURLPruner struct {
	keepAlt bool
}

func (p *ImageURLPruner) Transform(input string) (string, error) {
	// parse input as HTML, find all <img> tags, and remove their src attributes
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for i, a := range n.Attr {
				if a.Key == "src" {
					n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	var buf strings.Builder
	html.Render(&buf, doc)
	return buf.String(), nil
}

func NewImageURLPruner(keepAlt bool) *ImageURLPruner {
	return &ImageURLPruner{keepAlt: keepAlt}
}

type ClassPruner struct{}

func (p *ClassPruner) Transform(input string) (string, error) {
	// parse input as HTML, find all elements with class attributes, and remove the class attributes
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, a := range n.Attr {
				if a.Key == "class" {
					n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	var buf strings.Builder
	html.Render(&buf, doc)
	return buf.String(), nil
}

func NewClassPruner() *ClassPruner {
	return &ClassPruner{}
}

type StylePruner struct{}

func (p *StylePruner) Transform(input string) (string, error) {
	// parse input as HTML, find all elements with style attributes, and remove the style attributes
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", err
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, a := range n.Attr {
				if a.Key == "style" {
					n.Attr = append(n.Attr[:i], n.Attr[i+1:]...)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	var buf strings.Builder
	html.Render(&buf, doc)
	return buf.String(), nil
}

func NewStylePruner() *StylePruner {
	return &StylePruner{}
}
