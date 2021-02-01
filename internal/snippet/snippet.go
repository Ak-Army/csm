package snippet

import (
	"encoding/gob"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/juju/errors"

	"github.com/Ak-Army/csm/internal/config"
)

type List map[string]*Snippet

type Snippet struct {
	ID          int
	Title       string
	FileName    string
	Description string
	Username    string
	Name        string
	UpdatedAt   time.Time
	CreatedAt   time.Time
	Content     string
	Removable   bool
	Upgradeable bool
	config      *config.Config
}

func New() (*Snippet, error) {
	c, err := config.Get().Config()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Snippet{
		UpdatedAt: now,
		CreatedAt: now,
		config:    c,
	}, nil
}

func Get(fileName string) (*Snippet, error) {
	s, err := New()
	if err != nil {
		return nil, err
	}
	snippets, err := s.list()
	if err != nil {
		return nil, err
	}
	if snippet, ok := snippets[fileName]; ok {
		snippet.config = s.config
		return snippet, nil
	}
	return nil, errors.New("snippet not found")
}

func NewList() (List, error) {
	s, err := New()
	if err != nil {
		return nil, err
	}
	list, err := s.list()
	if err != nil {
		return nil, err
	}
	for _, snippet := range list {
		snippet.config = s.config
	}
	return list, nil
}

func (s *Snippet) CreateSnippet() (err error) {
	dataFilePath := s.snippetFilePath()
	if _, err = os.Stat(dataFilePath); !os.IsNotExist(err) {
		err = errors.New("snippet already exists")
		return
	}
	if _, err = os.Create(dataFilePath); err != nil {
		err = errors.Annotate(err, "snippet already exists")
		return
	}
	defer func() {
		if err != nil {
			os.Remove(dataFilePath)
		}
	}()
	dataFilePathTmp := dataFilePath + ".tmp"
	if err = ioutil.WriteFile(dataFilePathTmp, s.template(), os.ModePerm); err != nil {
		err = errors.Annotate(err, "unable to write data file")
		return
	}
	defer os.Remove(dataFilePathTmp)
	if err = s.edit(dataFilePathTmp); err != nil {
		err = errors.Annotate(err, "unable to create new snippet")
		return
	}
	if u, _ := user.Current(); u != nil {
		s.Username = u.Username
		s.Name = u.Name
	} else {
		s.Username = "unknown"
		s.Name = "Unknown"
	}
	s.Upgradeable = true
	err = s.Save()
	return
}

func (s *Snippet) Edit() error {
	dataFilePath := s.snippetFilePath()
	if _, err := os.Stat(dataFilePath); err != nil {
		return errors.New("snippet already exists")
	}
	dataFilePathTmp := dataFilePath + ".tmp"
	if err := ioutil.WriteFile(dataFilePathTmp, s.template(), os.ModePerm); err != nil {
		return errors.Annotate(err, "unable to write data file")
	}
	defer os.Remove(dataFilePathTmp)
	if err := s.edit(dataFilePathTmp); err != nil {
		return errors.Annotate(err, "unable to edit snippet")
	}
	s.Upgradeable = true
	s.UpdatedAt = time.Now()
	return s.Save()
}

func (s *Snippet) Remove() (err error) {
	snippets, err2 := s.list()
	if err2 != nil {
		err = err2
		return
	}
	if s.Removable {
		if err = os.Remove(s.snippetFilePath()); err != nil {
			err = errors.Annotate(err, "unable to remove snippet file")
			return
		}
		delete(snippets, s.FileName)
	}
	s.Removable = true
	err = s.saveDb(snippets)
	return
}

func (s *Snippet) Save() (err error) {
	dataFilePath := s.snippetFilePath()
	if err := ioutil.WriteFile(dataFilePath, []byte(s.Content), os.ModePerm); err != nil {
		return errors.Annotate(err, "unable to write data file")
	}
	defer func() {
		if err != nil {
			os.Remove(dataFilePath)
		}
	}()
	var snippets List
	snippets, err = s.list()
	if err != nil {
		return
	}
	snippets[s.FileName] = s
	err = s.saveDb(snippets)
	return
}

func (s *Snippet) list() (List, error) {
	snippets := make(List)
	dataFile, err := os.Open(s.config.DbFilePath())
	if os.IsNotExist(err) {
		return snippets, nil
	}
	if err != nil {
		return snippets, errors.Annotate(err, "unable to open database file")
	}
	defer dataFile.Close()
	dataDecoder := gob.NewDecoder(dataFile)
	if err = dataDecoder.Decode(&snippets); err != nil {
		return snippets, errors.Annotate(err, "unable to decode database file")
	}
	return snippets, nil
}

func (s *Snippet) template() []byte {
	var tmp []string
	delim := "-----------"
	if s.Title != "" {
		tmp = append(tmp, s.Title)
	} else {
		tmp = append(tmp, "Title")
	}
	tmp = append(tmp, delim)
	if s.Description != "" {
		tmp = append(tmp, s.Description)
	} else {
		tmp = append(tmp, "Description")
	}
	tmp = append(tmp, delim)
	if s.Content != "" {
		tmp = append(tmp, s.Content)
	} else {
		tmp = append(tmp, "Snippet content")
	}
	return []byte(strings.Join(tmp, "\n"))
}

func (s *Snippet) edit(datafile string) error {
	cmd := exec.Command(s.config.Editor, datafile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
	bs, err := ioutil.ReadFile(datafile)
	if err != nil {
		return err
	}
	ss := strings.Split(string(bs), "\n")
	if len(ss) < 4 {
		return errors.New("Not enough data")
	}
	i := 0
	s.Title = ss[i]
	i++
	delim := ss[i]
	i++
	var descprition []string
	for ; i < len(ss); i++ {
		if ss[i] == delim {
			i++
			break
		}
		descprition = append(descprition, ss[i])
	}
	s.Description = strings.Join(descprition, "\n")
	var content []string
	for ; i < len(ss); i++ {
		content = append(content, ss[i])
	}
	s.Content = strings.Join(content, "\n")
	return nil
}

func (s *Snippet) saveDb(snippets List) (err error) {
	if _, ok := snippets[s.FileName]; ok {
		snippets[s.FileName] = s
	}
	dbFile, err2 := os.Create(s.config.DbFilePath())
	if err2 != nil {
		err = errors.Annotate(err2, "unable to open database file")
		return
	}
	dataEncoder := gob.NewEncoder(dbFile)
	if err = dataEncoder.Encode(snippets); err != nil {
		err = errors.Annotate(err, "unable to decode database file")
		return
	}
	return nil
}

func (s *Snippet) snippetFilePath() string {
	return filepath.Join(s.config.SnippetPath, s.FileName)
}
