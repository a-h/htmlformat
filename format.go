package htmlformat

import (
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Document formats a HTML document.
func Document(w io.Writer, r io.Reader) (err error) {
	node, err := html.Parse(r)
	if err != nil {
		return err
	}
	return Nodes(w, []*html.Node{node})
}

// Fragment formats a fragment of a HTML document.
func Fragment(w io.Writer, r io.Reader) (err error) {
	context := &html.Node{
		Type: html.ElementNode,
	}
	nodes, err := html.ParseFragment(r, context)
	if err != nil {
		return err
	}
	return Nodes(w, nodes)
}

// Nodes formats a slice of HTML nodes.
func Nodes(w io.Writer, nodes []*html.Node) (err error) {
	for _, node := range nodes {
		if err = printNode(w, node, false, 0); err != nil {
			return
		}
	}
	return
}

// The <pre> tag indicates that the text within it should always be formatted
// as is. See https://github.com/ericchiang/pup/issues/33
func printPre(w io.Writer, n *html.Node) (err error) {
	switch n.Type {
	case html.TextNode:
		s := n.Data
		if _, err = fmt.Fprint(w, s); err != nil {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err = printPre(w, c); err != nil {
				return
			}
		}
	case html.ElementNode:
		if _, err = fmt.Fprintf(w, "<%s", n.Data); err != nil {
			return
		}
		for _, a := range n.Attr {
			val := html.EscapeString(a.Val)
			if _, err = fmt.Fprintf(w, ` %s="%s"`, a.Key, val); err != nil {
				return
			}
		}
		if _, err = fmt.Fprint(w, ">"); err != nil {
			return
		}
		if !isVoidElement(n) {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if err = printPre(w, c); err != nil {
					return
				}
			}
			if _, err = fmt.Fprintf(w, "</%s>", n.Data); err != nil {
				return
			}
		}
	case html.CommentNode:
		data := n.Data
		if _, err = fmt.Fprintf(w, "<!--%s-->\n", data); err != nil {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err = printPre(w, c); err != nil {
				return
			}
		}
	case html.DoctypeNode, html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if err = printPre(w, c); err != nil {
				return
			}
		}
	}
	return
}

// Is this node a tag with no end tag such as <meta> or <br>?
// http://www.w3.org/TR/html-markup/syntax.html#syntax-elements
func isVoidElement(n *html.Node) bool {
	switch n.DataAtom {
	case atom.Area, atom.Base, atom.Br, atom.Col, atom.Command, atom.Embed,
		atom.Hr, atom.Img, atom.Input, atom.Keygen, atom.Link,
		atom.Meta, atom.Param, atom.Source, atom.Track, atom.Wbr:
		return true
	}
	return false
}

func getFirstRune(s string) rune {
	r, _ := utf8.DecodeRuneInString(s)
	return r
}

func hasSingleTextChild(n *html.Node) bool {
	return n != nil && n.FirstChild != nil && n.FirstChild == n.LastChild && n.FirstChild.Type == html.TextNode
}

func printNode(w io.Writer, n *html.Node, pre bool, level int) (err error) {
	switch n.Type {
	case html.TextNode:
		if pre {
			if _, err = fmt.Fprint(w, n.Data); err != nil {
				return
			}
			return nil
		}
		s := n.Data
		s = strings.TrimSpace(s)
		if s != "" {
			if !hasSingleTextChild(n.Parent) &&
				(n.PrevSibling == nil || !unicode.IsPunct(getFirstRune(s))) {
				if err = printIndent(w, level); err != nil {
					return
				}
			}
			if _, err = fmt.Fprint(w, s); err != nil {
				return
			}
			if !hasSingleTextChild(n.Parent) {
				if _, err = fmt.Fprint(w, "\n"); err != nil {
					return
				}
			}
		}
	case html.ElementNode:
		if err = printIndent(w, level); err != nil {
			return
		}
		if _, err = fmt.Fprintf(w, "<%s", n.Data); err != nil {
			return
		}
		for _, a := range n.Attr {
			val := html.EscapeString(a.Val)
			if _, err = fmt.Fprintf(w, ` %s="%s"`, a.Key, val); err != nil {
				return
			}
		}
		terminator := "\n"
		if isPreFormatted(n.Data) {
			terminator = ""
		}
		if _, err = fmt.Fprint(w, ">", terminator); err != nil {
			return
		}
		if !hasSingleTextChild(n) {
			if _, err = fmt.Fprint(w, "\n"); err != nil {
				return
			}
		}
		if !isVoidElement(n) {
			if err = printChildren(w, n, isPreFormatted(n.Data), level+1); err != nil {
				return
			}
			if !hasSingleTextChild(n) {
				if err = printIndent(w, level); err != nil {
					return
				}
			}
			if _, err = fmt.Fprintf(w, "</%s>", n.Data); err != nil {
				return
			}

			if n.NextSibling == nil ||
				(!unicode.IsPunct(getFirstRune(n.NextSibling.Data)) || n.NextSibling.Type == html.ElementNode) {
				if _, err = fmt.Fprint(w, "\n"); err != nil {
					return
				}
			}
		}
	case html.CommentNode:
		if err = printIndent(w, level); err != nil {
			return
		}
		if _, err = fmt.Fprintf(w, "<!--%s-->\n", n.Data); err != nil {
			return
		}
		if err = printChildren(w, n, false, level); err != nil {
			return
		}
	case html.DoctypeNode, html.DocumentNode:
		if err = printChildren(w, n, false, level); err != nil {
			return
		}
	}
	return
}

func isPreFormatted(s string) bool {
	return s == "pre" || s == "script" || s == "style"
}

func printChildren(w io.Writer, n *html.Node, pre bool, level int) (err error) {
	child := n.FirstChild
	for child != nil {
		if err = printNode(w, child, pre, level); err != nil {
			return
		}
		child = child.NextSibling
	}
	return
}

func printIndent(w io.Writer, level int) (err error) {
	_, err = fmt.Fprint(w, strings.Repeat(" ", level))
	return err
}
