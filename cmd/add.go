package cmd

import (
	"strings"

	"github.com/Ak-Army/csm/internal/snippet"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"
)

type Add struct {
	*cli.Flagger
	Command string `flag:"command, the default command"`
}

func (a *Add) Desc() string {
	return "Create a new snippet"
}
func (a *Add) Run() {
	s, err := snippet.New()
	if err != nil {
		xlog.Fatal("Unable to create snippet", err)
	}
	s.Content = a.Command
	args := a.Args()
	xlog.Info("Add new snippet ", strings.Join(a.Args(), " "))
	if len(args) != 1 {
		xlog.Fatal("Not enough arguments")
	}
	s.FileName = args[0]
	if err := s.CreateSnippet(); err != nil {
		xlog.Fatal("Unable to save snippet", err)
	}
}
