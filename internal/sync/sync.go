package sync

import (
	"github.com/Ak-Army/csm/internal/snippet"
)

type Client interface {
	GetSnippets() (snippet.List, error)
	UploadSnippet(s *snippet.Snippet) error
	DeleteSnippet(s *snippet.Snippet) error
	GetContent(id int) (string, error)
}
