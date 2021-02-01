package cmd

import (
	"fmt"

	"github.com/Ak-Army/csm/internal/snippet"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"
	"github.com/mgutz/ansi"
)

type List struct {
	*cli.Flagger
	Full bool `flag:"full, show every data from snippet"`
}

func (l *List) Desc() string {
	return "List all snippet"
}
func (l *List) Run() {
	snippets, err := snippet.NewList()
	if err != nil {
		xlog.Fatal("Unable to create snippet", err)
	}
	i := 0
	for _, s := range snippets {
		if s.Removable && !l.Full {
			continue
		}
		i++
		fmt.Println(
			ansi.Color(fmt.Sprintf("%d.", i), "+b:yellow"),
			ansi.Color(s.Title, "+b"),
			fmt.Sprintf("(ID: %d)", s.ID),
		)
		if l.Full {
			if s.Removable {
				fmt.Println(ansi.Color("removable", "red+b"))
			}
			fmt.Println(
				ansi.Color("User:", "blue"),
				fmt.Sprintf("%s (%s)", s.Name, s.Username),
				ansi.Color("Created at:", "gray"),
				s.CreatedAt,
				ansi.Color("Updated at:", "gray"),
				s.UpdatedAt,
			)
			if s.Description != "" {
				fmt.Println(s.Description)
			}
		}
		fmt.Println(
			ansi.Color("Command:", "green"),
			fmt.Sprintf("(%s)", s.FileName),
		)
		fmt.Println(s.Content)
		fmt.Println("-------------------------------------------")
	}

}
