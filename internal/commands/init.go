package commands

import "github.com/theoreotm/gommand"

var cmds []*gommand.Command

var infoCategory = &gommand.Category{
	Name:        "Information",
	Description: "General commands to retrieve info.",
}

func Register(router *gommand.Router) {
	for _, v := range cmds {
		router.SetCommand(v)
	}
}
