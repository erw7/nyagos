package texts

import (
	"strings"
	"unicode"
)

const nulquote = '\000'

func quotedWordCutter(reader *strings.Reader) (string, bool) {
	var buffer strings.Builder
	for {
		if reader.Len() <= 0 {
			return "", false
		}
		ch, _, _ := reader.ReadRune()
		if !unicode.IsSpace(ch) {
			reader.UnreadRune()
			break
		}
	}
	quote := nulquote
	yenCount := 0
	for reader.Len() > 0 {
		ch, _, _ := reader.ReadRune()
		if yenCount%2 == 0 {
			if quote == nulquote && (ch == '"' || ch == '\'') {
				quote = ch
			} else if quote != nulquote && ch == quote {
				quote = nulquote
			}
		}
		if unicode.IsSpace(ch) && quote == nulquote {
			break
		}
		if ch == '\\' {
			yenCount++
		} else {
			yenCount = 0
		}
		buffer.WriteRune(ch)
	}
	return buffer.String(), true
}

// SplitLikeShellString - Split s with SPACES not enclosing with double-quotations.
func SplitLikeShellString(line string) []string {
	args := make([]string, 0, 10)
	reader := strings.NewReader(line)
	for reader.Len() > 0 {
		word, ok := quotedWordCutter(reader)
		if ok {
			args = append(args, word)
		}
	}
	return args
}

func FirstWord(line string) string {
	reader := strings.NewReader(line)
	str, _ := quotedWordCutter(reader)
	return str
}
