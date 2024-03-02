package handler

import "strings"

type Argument struct {
	Name        string
	Description string
	Type        int
	Required    bool
}

func ParseArgs(input string) []string {
	return strings.Fields(input)
}
