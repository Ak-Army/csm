package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Ak-Army/csm/internal/snippet"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/mgutz/ansi"
)

type Query struct {
	*cli.Flagger
	Max int `flag:"max, maximum result"`
}

type snippetWithDistance struct {
	*snippet.Snippet
	Distance int
}

func (q *Query) Desc() string {
	return "List all snippet"
}
func (q *Query) Run() {
	snippets, err := snippet.NewList()
	if err != nil {
		xlog.Fatal("Unable to list snippets", err)
	}
	args := q.Args()
	xlog.Info("Query snippets for ", strings.Join(args, " "))
	if len(args) != 1 {
		xlog.Fatal("Not enough arguments")
	}
	var found orderedSnippets
	for _, s := range snippets {
		if s.Removable {
			continue
		}
		ranks := fuzzy.RankFindFold(strings.Join(args, " "), []string{s.Title, s.Description})
		if ranks.Len() > 0 {
			sort.Sort(ranks)
			found = append(found, snippetWithDistance{
				Snippet:  s,
				Distance: ranks[0].Distance,
			})
		}
	}

	sort.Sort(found)
	i := 0
	for _, s := range found {
		i++
		if i > q.Max {
			break
		}
		fmt.Println(
			ansi.Color(fmt.Sprintf("%d.", i), "+b:yellow"),
			ansi.Color(s.Title, "+b"),
		)
		fmt.Println(
			ansi.Color("Command:", "green"),
			fmt.Sprintf("(%s)", s.FileName),
		)
		fmt.Println(s.Content)
		fmt.Println("-------------------------------------------")
	}

}

type orderedSnippets []snippetWithDistance

func (r orderedSnippets) Len() int {
	return len(r)
}

func (r orderedSnippets) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r orderedSnippets) Less(i, j int) bool {
	return r[i].Distance < r[j].Distance
}
