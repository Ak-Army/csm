package cmd

import (
	"strings"

	"github.com/Ak-Army/csm/internal/snippet"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"
)

type Rm struct {
	*cli.Flagger
}

func (r *Rm) Desc() string {
	return "Remove a snippet"
}
func (r *Rm) Run() {
	args := r.Args()
	xlog.Info("Remove snippet ", strings.Join(r.Args(), " "))
	if len(args) != 1 {
		xlog.Fatal("Not enough arguments")
	}
	s, err := snippet.Get(args[0])
	if err != nil {
		xlog.Fatal("Unable to remove snippet", err)
	}
	if err := s.Remove(); err != nil {
		xlog.Fatal("Unable to remove snippet", err)
	}
}
