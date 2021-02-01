package cmd

import (
	"time"

	"github.com/Ak-Army/csm/internal/snippet"
	"github.com/Ak-Army/csm/internal/sync/gitlab"

	"github.com/Ak-Army/cli"
	"github.com/Ak-Army/xlog"
)

type Sync struct {
	*cli.Flagger
}

func (s *Sync) Desc() string {
	return "Create s new snippet"
}
func (s *Sync) Run() {
	snippets, err := snippet.NewList()
	if err != nil {
		xlog.Fatal("Unable to create snippet", err)
	}
	gitLabClient, err := gitlab.New()
	if err != nil {
		xlog.Fatal("Unable to initialize gitlab", err)
	}
	syncStat := struct {
		upload   int
		update   int
		add      int
		remove   int
		download int
	}{}
	for _, s := range snippets {
		if s.Upgradeable == true {
			if s.ID != 0 {
				syncStat.upload++
			} else {
				syncStat.add++
			}
			if err := gitLabClient.UploadSnippet(s); err != nil {
				xlog.Fatal("Unable to sync updated snippet", err)
			}
			s.Upgradeable = false
			s.UpdatedAt = time.Now()
			if err := s.Save(); err != nil {
				xlog.Fatal("Unable to save snippet", err)
			}
		}
		if s.Removable {
			syncStat.remove++
			xlog.Info("Remove snippet: ", s.ID)
			if err := gitLabClient.DeleteSnippet(s); err != nil {
				xlog.Fatal("Unable to sync removable snippet", err)
			}
			if err := s.Remove(); err != nil {
				xlog.Fatal("Unable to remove snippet", err)
			}
		}
	}
	gitlabSnippets, err := gitLabClient.GetSnippets()
	if err != nil {
		xlog.Fatal("Unable to sync snippets", err)
	}
	for id, s := range gitlabSnippets {
		if _, ok := snippets[id]; !ok {
			syncStat.download++
			xlog.Info("Download snippet: ", s.ID)
			if s.Content, err = gitLabClient.GetContent(s.ID); err != nil {
				xlog.Fatal("Unable to get snippet content", err)
			}
			if err := s.Save(); err != nil {
				xlog.Fatal("Unable to save snippet", err)
			}
		} else if ss, ok := snippets[id]; ok && ss.UpdatedAt.Before(s.UpdatedAt) {
			syncStat.update++
			if s.Content, err = gitLabClient.GetContent(s.ID); err != nil {
				xlog.Fatal("Unable to get snippet content", err)
			}
			if err := s.Save(); err != nil {
				xlog.Fatal("Unable to save snippet", err)
			}
		}
	}
	xlog.Infof("Sync done %+v", syncStat)
}
