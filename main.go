package main

import (
	"os"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"

	"github.com/Ak-Army/csm/cmd"
	"github.com/Ak-Army/csm/internal/config"
)

func main() {
	_, err := config.Get().Config()
	if err != nil {
		xlog.Fatal(err)
	}
	c := cli.New("csm", Version+" "+BuildTime)
	c.Authors = []string{"authors goes here"}
	c.Add(
		&cmd.Add{},
		&cmd.Rm{},
		&cmd.Edit{},
		&cmd.Sync{},
		&cmd.List{},
		&cmd.Query{
			Max: 10,
		},
	)
	c.Run(os.Args)
}
