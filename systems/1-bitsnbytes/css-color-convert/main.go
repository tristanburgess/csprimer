package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"
)

var (
	srcPath = flag.String("src_path", "", "path to source .css file")
)

type tokenKind int

const (
	tokBodyIdentBegin tokenKind = iota
	tokBodyIdentEnd
	tokCommentBegin
	tokCommentBody
	tokCommentEnd
	tokTagIdent
	tokValHexColor3or6
	tokValHexColor4or8
)

func (s tokenKind) String() string {
	switch s {
	case tokBodyIdentBegin:
		return "tokBodyIdentBegin"
	case tokBodyIdentEnd:
		return "tokBodyIdentEnd"
	case tokCommentBegin:
		return "tokCommentBegin"
	case tokCommentBody:
		return "tokCommentBody"
	case tokCommentEnd:
		return "tokCommentEnd"
	case tokTagIdent:
		return "tokTagIdent"
	case tokValHexColor3or6:
		return "tokValHexColor3or6"
	case tokValHexColor4or8:
		return "tokValHexColor4or8"
	default:
		return "unknown"
	}
}

type token struct {
	Kind tokenKind
	val  string
}

func hexToRGB(t token) token {
	fmt.Printf("\ntest: %v\n", t)
	return t
}

func hexToRGBA(t token) token {
	fmt.Printf("\ntest: %v\n", t)
	return t
}

func tokenize(s string) ([]token, error) {
	toks := make([]token, 0)
	s = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)

	curVal := ""
	commentInFlight := false
	hexValLen := -1
	for i, c := range s {
		if commentInFlight && c != '*' && c != '/' {
			curVal += string(c)
			continue
		}

		switch c {
		case '/':
			{
				curVal += string(c)
				if len(curVal) >= 2 && curVal[len(curVal)-2:] == "*/" {
					commentInFlight = false
					toks = append(toks, token{
						Kind: tokCommentBody,
						val:  curVal[:len(curVal)-2],
					}, token{
						Kind: tokCommentEnd,
						val:  "*/",
					})
					curVal = ""
				}
			}
		case '*':
			{
				curVal += string(c)
				if curVal == "/*" {
					commentInFlight = true
					toks = append(toks, token{
						Kind: tokCommentBegin,
						val:  curVal,
					})
					curVal = ""
				}
			}
		case '{':
			{
				curVal += string(c)
				toks = append(toks, token{
					Kind: tokBodyIdentBegin,
					val:  curVal,
				})
				curVal = ""
			}
		case '}':
			{
				toks = append(toks, token{
					Kind: tokBodyIdentEnd,
					val:  curVal,
				})
				curVal = ""
			}
		case ':':
			{
				curVal += string(c)
				toks = append(toks, token{
					Kind: tokTagIdent,
					val:  curVal,
				})
				curVal = ""
			}
		case '#':
			{
				curVal += string(c)
				hexValLen = 0
			}
		case ';':
			{
				curVal += string(c)
				if hexValLen < 0 {
					return nil, fmt.Errorf(
						"hit unknown value kind found ending at pos: %d while building cur value: %#v",
						i, curVal,
					)
				}

				var k tokenKind
				if hexValLen == 3 || hexValLen == 6 {
					k = tokValHexColor3or6
				} else if hexValLen == 4 || hexValLen == 8 {
					k = tokValHexColor4or8
				} else {
					return nil, fmt.Errorf(
						"hit invalid length hex value found ending at pos: %d while building cur value: %#v",
						i, curVal,
					)
				}
				regex, err := regexp.Compile(fmt.Sprintf("#([0-9]|[a-f]){%d};", hexValLen))
				if err != nil {
					return nil, err
				}
				if !regex.MatchString(curVal) {
					return nil, fmt.Errorf(
						"hit invalid hex value found ending at pos: %d while building cur value: %#v",
						i, curVal,
					)
				}

				toks = append(toks, token{
					Kind: k,
					val:  curVal,
				})
				curVal = ""
				hexValLen = -1
			}
		case '0', '1', '2', '3', '4', '5',
			'6', '7', '8', '9',
			'a', 'b', 'c', 'd', 'e', 'f':
			{
				curVal += string(c)
				if hexValLen >= 0 {
					hexValLen += 1
				}
			}
		case 'A', 'B', 'C', 'D', 'E', 'F':
			{
				if hexValLen >= 0 {
					curVal += strings.ToLower(string(c))
					hexValLen += 1
				} else {
					curVal += string(c)
				}
			}
		case '.', '-', 'g', 'G', 'h', 'H',
			'i', 'I', 'j', 'J', 'k', 'K',
			'l', 'L', 'm', 'M', 'n', 'N',
			'o', 'O', 'p', 'P', 'q', 'Q',
			'r', 'R', 's', 'S', 't', 'T',
			'u', 'U', 'v', 'V', 'w', 'W',
			'x', 'X', 'y', 'Y', 'z', 'Z':
			{
				curVal += string(c)
			}
		default:
			{
				return nil, fmt.Errorf(
					"hit unexpected string constant: %#v at pos: %d while building cur value: %#v",
					string(c), i, curVal,
				)
			}
		}
	}

	return toks, nil
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	file, err := os.Open(*srcPath)
	check(err)
	bs, err := io.ReadAll(file)
	check(err)

	srcText := string(bs)
	fmt.Fprintf(os.Stderr, "src code text: %q\n", srcText)
	toks, err := tokenize(srcText)
	check(err)
	fmt.Fprintf(os.Stderr, "src toks: %q\n", toks)
	// WARNING: no parsing implemented past tokenization.
	// assumes source program tokens form a valid syntax.
	for i, tok := range toks {
		if tok.Kind == tokValHexColor3or6 {
			toks[i] = hexToRGB(tok)
		} else if tok.Kind == tokValHexColor4or8 {
			toks[i] = hexToRGBA(tok)
		}
	}
	fmt.Fprintf(os.Stderr, "xformed toks: %q", toks)
}
