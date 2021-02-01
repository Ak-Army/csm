package cmd

import (
	"strings"

	"github.com/Ak-Army/csm/internal/snippet"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"
)

type Edit struct {
	*cli.Flagger
	Command string `flag:"command, the default command"`
}

func (e *Edit) Desc() string {
	return "Edit a snippet"
}
func (e *Edit) Run() {
	args := e.Args()
	xlog.Info("Edit snippet ", strings.Join(e.Args(), " "))
	if len(args) != 1 {
		xlog.Fatal("Not enough arguments")
	}
	s, err := snippet.Get(args[0])
	if err != nil {
		xlog.Fatal("Unable to get snippet", err)
	}
	xlog.Infof("Edit snippet %#v", s)
	if err := s.Edit(); err != nil {
		xlog.Fatal("Unable to edit snippet", err)
	}
}
