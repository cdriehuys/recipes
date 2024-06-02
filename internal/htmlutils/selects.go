package htmlutils

import (
	"errors"
	"fmt"

	"golang.org/x/net/html"
)

// ErrNodeNotFound indicates that no HTML node matching the provided query was found.
var ErrNodeNotFound = errors.New("node not found")

// Select contains the simplified representation of a <select> tag.
type Select struct {
	// Options contains a representation of the <select> tag's options.
	Options []Option
}

// Option is a simplified representation of an <option> tag for a <select>.
type Option struct {
	// Selected indicates if the tag had the "select" attribute.
	Selected bool
	// Value is the "value" attribute of the tag.
	Value string
}

// FindSelectInput finds a <select> input by name from the provided HTML node.
func FindSelectInput(node *html.Node, name string) (Select, error) {
	selectNode := findSelectByName(node, name)
	if selectNode == nil {
		return Select{}, fmt.Errorf("no select node named %s: %w", name, ErrNodeNotFound)
	}

	selectTag := Select{
		Options: collectOptions(selectNode),
	}

	return selectTag, nil
}

func findSelectByName(node *html.Node, name string) *html.Node {
	if node.Type == html.ElementNode && node.Data == "select" {
		for _, attr := range node.Attr {
			if attr.Key == "name" && attr.Val == name {
				return node
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if foundChild := findSelectByName(child, name); foundChild != nil {
			return foundChild
		}
	}

	return nil
}

func collectOptions(selectNode *html.Node) []Option {
	var options []Option

	for child := selectNode.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "option" {
			opt := Option{}
			for _, attr := range child.Attr {
				if attr.Key == "value" {
					opt.Value = attr.Val
				}

				if attr.Key == "selected" {
					opt.Selected = true
				}
			}

			options = append(options, opt)
		}
	}

	return options
}
