package synonym

import (
	"fmt"
	"regexp"
	"strings"
)

var synonyms map[string]string

func init() {
	synonyms = map[string]string{}
}

func Add(word, synonym string) *map[string]string {
	if synonyms == nil {
		synonyms = map[string]string{}
	}
	synonyms[word] = synonym
	synonyms[synonym] = word
	return &synonyms
}

func regex() *regexp.Regexp {
	keys := make([]string, len(synonyms))
	i := 0
	for k := range synonyms {
		keys[i] = k
		i++
	}

	r := `(?i)\b(` + strings.Join(keys, "|") + `)\b`
	return regexp.MustCompile(r)
}

func ReplaceAll(search string) string {
	replaced := regex().ReplaceAllFunc([]byte(search), func(match []byte) []byte {
		v, ok := synonyms[string(match)]
		if ok {
			return []byte(fmt.Sprintf("(%s OR %s)", match, v))
		}
		return match
	})
	return string(replaced)
}
