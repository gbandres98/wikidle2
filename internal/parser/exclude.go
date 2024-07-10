package parser

var ExcludedWords = []string{
	"a",
	"con",
	"de",
	"del",
	"en",
	"para",
	"por",
	"sin",
	"el",
	"la",
	"los",
	"las",
	"un",
	"uno",
	"unos",
	"una",
	"unas",
	"y",
	"o",
	"u",
	"e",
	"que",
	"le",
	"les",
	"lo",
	"los",
}

func IsExcludedWord(word string) bool {
	word = Normalize(word)
	for _, excludedWord := range ExcludedWords {
		if word == excludedWord {
			return true
		}
	}

	return false
}
