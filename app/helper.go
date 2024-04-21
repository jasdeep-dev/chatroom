package app

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Titleize(name string) string {
	caser := cases.Title(language.English)
	return caser.String(name)
}
