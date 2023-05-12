package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
)

var (
	srcPath = flag.String("src_path", "./tests/advanced.css", "path to source .css file")
)

type tokenKind int

const (
	tokOther tokenKind = iota
	tokValHexColor
	tokValRGBColor
)

func (s tokenKind) String() string {
	switch s {
	case tokOther:
		return "tokOther"
	case tokValHexColor:
		return "tokValHexColor"
	case tokValRGBColor:
		return "tokValRGBColor"
	default:
		return "unknown"
	}
}

type token struct {
	Kind    tokenKind
	dataLen int
	rawVal  string
}

func hexToRGBA(t token) token {
	t.Kind = tokValRGBColor
	t.rawVal = t.rawVal[1 : len(t.rawVal)-1]
	rgba := make([]byte, 4)
	aStr := ""
	digits := 3
	if (len(t.rawVal))%4 == 0 {
		aStr = "a"
		digits = 4
	}
	digitsPerChannel := len(t.rawVal) / digits
	for i := 0; i < len(t.rawVal); i++ {
		idx := i / (digitsPerChannel)
		// digitsPerChannel is only ever 1 or 2.
		// If digitsPerChannel is 1, we adhere to the extension rule to copy
		// the channel digit X to be interpreted as hex value XX.
		// Otherwise, digitsPerChannel must be 2 and so we process
		// each digit only once.
		for j := 0; j < 3-digitsPerChannel; j++ {
			rgba[idx] <<= 4
			if t.rawVal[i] >= 'a' {
				rgba[idx] += 10 + byte(t.rawVal[i]-'a')
			} else {
				rgba[idx] += byte(t.rawVal[i] - '0')
			}
		}
	}
	rgbStr := make([]string, 3)
	for i := 0; i < 3; i++ {
		if rgba[i] > 0 && rgba[3] > 0 {
			rgbStr[i] = fmt.Sprintf("%d / %.05f", rgba[i], float32(rgba[3])/255.0)
		} else {
			rgbStr[i] = fmt.Sprint(rgba[i])
		}
	}
	t.rawVal = fmt.Sprintf("rgb%s(%s %s %s);", aStr, rgbStr[0], rgbStr[1], rgbStr[2])
	return t
}

func tokenize(s string) ([]token, error) {
	toks := make([]token, 0)

	curVal := ""
	hexValLen := -1
	for i, c := range s {
		switch c {
		case '#':
			{
				if len(curVal) > 0 {
					toks = append(toks, token{
						Kind:   tokOther,
						rawVal: curVal,
					})
					curVal = ""
				}
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

				if hexValLen != 3 && hexValLen != 6 && hexValLen != 4 && hexValLen != 8 {
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
					Kind:    tokValHexColor,
					dataLen: hexValLen,
					rawVal:  curVal,
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
					curVal += string(s[i] + ('a' - 'A'))
					hexValLen += 1
				} else {
					curVal += string(c)
				}
			}
		default:
			{
				curVal += string(c)
			}
		}
	}

	if len(curVal) > 0 {
		toks = append(toks, token{
			Kind:   tokOther,
			rawVal: curVal,
		})
		curVal = ""
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
	fmt.Fprintf(os.Stderr, "src code text: %q\n\n", srcText)
	toks, err := tokenize(srcText)
	check(err)

	fmt.Fprintf(os.Stderr, "src toks: %q\n\n", toks)
	for i, tok := range toks {
		if tok.Kind == tokValHexColor {
			toks[i] = hexToRGBA(tok)
		}
	}
	fmt.Fprintf(os.Stderr, "xformed toks: %q\n\n", toks)
	xformedSrc := ""
	for _, tok := range toks {
		xformedSrc += fmt.Sprint(tok.rawVal)
	}
	fmt.Fprintf(os.Stderr, "xformed src code text: %q\n", xformedSrc)
	fmt.Fprint(os.Stderr, "\n")

	fmt.Print(xformedSrc)
}
