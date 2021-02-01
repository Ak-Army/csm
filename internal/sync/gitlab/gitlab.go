package gitlab

import (
	"time"

	"github.com/juju/errors"
	"github.com/xanzy/go-gitlab"

	"github.com/Ak-Army/csm/internal/config"
	"github.com/Ak-Army/csm/internal/snippet"
	"github.com/Ak-Army/csm/internal/sync"
)

type GitLab struct {
	Client *gitlab.Client
	config *config.Config
}

func New() (sync.Client, error) {
	c, err := config.Get().Config()
	if err != nil {
		return nil, err
	}
	if c.GitlabAccessToken == "" {
		return nil, errors.New(`access_token is empty.
Go https://gitlab.com/profile/personal_access_tokens and create access_token.
Write access_token in config file.
		`)
	}
	client := GitLab{
		Client: gitlab.NewClient(nil, c.GitlabAccessToken),
		config: c,
	}
	if c.GitlabURL != "" {
		if err := client.Client.SetBaseURL(c.GitlabURL); err != nil {
			return nil, err
		}
	}
	return client, nil
}

func (g GitLab) GetSnippets() (snippet.List, error) {
	gitlabSnippets, res, err := g.Client.Snippets.ListSnippets(&gitlab.ListSnippetsOptions{})
	if err != nil {
		if res != nil && res.StatusCode == 404 {
			return nil, errors.Annotate(err, "no GitLab Snippet found")
		}
		return nil, errors.Annotate(err, "failed to get GitLab Snippet")
	}
	snippets := make(snippet.List)
	for _, s := range gitlabSnippets {
		if s.FileName != "" {
			getTime := func(t *time.Time) time.Time {
				if t == nil {
					return time.Now()
				}
				return *t
			}
			ss, err := snippet.New()
			if err != nil {
				return snippets, err
			}
			ss.ID = s.ID
			ss.Title = s.Title
			ss.FileName = s.FileName
			ss.Description = s.Description
			ss.Username = s.Author.Username
			ss.Name = s.Author.Name
			ss.UpdatedAt = getTime(s.UpdatedAt)
			ss.CreatedAt = getTime(s.CreatedAt)
			snippets[s.FileName] = ss

		}
	}
	return snippets, nil
}

func (g GitLab) UploadSnippet(s *snippet.Snippet) error {
	if s.ID == 0 {
		id, err := g.createSnippet(s)
		if err != nil {
			return errors.Annotate(err, "failed to create GitLab Snippet")
		}
		s.ID = id
	} else {
		if err := g.updateSnippet(s); err != nil {
			return errors.Annotate(err, "failed to update GitLab Snippet")
		}
	}
	return nil
}

func (g GitLab) DeleteSnippet(s *snippet.Snippet) error {
	if s.ID != 0 {
		if _, err := g.Client.Snippets.DeleteSnippet(s.ID); err != nil {
			return errors.Annotate(err, "failed to delete GitLab Snippet")
		}
	}
	return nil
}

func (g GitLab) GetContent(id int) (string, error) {
	content, _, err := g.Client.Snippets.SnippetContent(id)
	if err != nil {
		return "", errors.Annotate(err, "failed to get snippet content from GitLab")
	}
	return string(content), nil
}

func (g GitLab) createSnippet(s *snippet.Snippet) (int, error) {
	opt := &gitlab.CreateSnippetOptions{
		Title:       gitlab.String(s.Title),
		FileName:    gitlab.String(s.FileName),
		Description: gitlab.String(s.Description),
		Content:     gitlab.String(s.Content),
		Visibility:  gitlab.Visibility(gitlab.PublicVisibility),
	}
	ret, _, err := g.Client.Snippets.CreateSnippet(opt)
	if err != nil {
		return -1, errors.Annotate(err, "failed to create GitLab Snippet")
	}
	return ret.ID, nil
}

func (g GitLab) updateSnippet(s *snippet.Snippet) error {
	opt := &gitlab.UpdateSnippetOptions{
		Title:       gitlab.String(s.Title),
		FileName:    gitlab.String(s.FileName),
		Description: gitlab.String(s.Description),
		Content:     gitlab.String(s.Content),
		Visibility:  gitlab.Visibility(gitlab.PublicVisibility),
	}
	if _, _, err := g.Client.Snippets.UpdateSnippet(s.ID, opt); err != nil {
		return errors.Annotate(err, "failed to update GitLab Snippet")
	}
	return nil
}
