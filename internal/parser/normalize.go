package parser

import "strings"

func Normalize(word string) string {
	word = strings.ToLower(word)
	word = strings.Trim(word, ".,;:()[]{}\"'¿?¡! ")
	word = strings.ReplaceAll(word, "á", "a")
	word = strings.ReplaceAll(word, "é", "e")
	word = strings.ReplaceAll(word, "í", "i")
	word = strings.ReplaceAll(word, "ó", "o")
	word = strings.ReplaceAll(word, "ú", "u")
	return word
}
