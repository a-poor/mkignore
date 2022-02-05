package main

import (
	"io"
	"path"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

const (
	RepoURL            = "https://github.com/github/gitignore"
	CommunityDirPrefix = "/community/"
	GlobalDirPrefix    = "/Global/"
	GitignoreExtension = ".gitignore"
)

type IgnoreFile struct {
	Name string
	Path string
	Data string
}

func (i IgnoreFile) SplitPath() []string {
	var dirs []string
	p, _ := path.Split(i.Path)
	for !(p == "" || p == "." || p == "/") {
		var d string
		p, d = path.Split(p)
		dirs = append(dirs, d)
	}
	return dirs
}

func (i IgnoreFile) IsCommunity() bool {
	return strings.HasPrefix(i.Path, CommunityDirPrefix)
}

func (i IgnoreFile) IsGlobal() bool {
	return strings.HasPrefix(i.Path, GlobalDirPrefix)
}

func (i IgnoreFile) GetLabel() string {
	s := i.Name
	if i.IsCommunity() {
		s += " (community)"
	}
	if i.IsGlobal() {
		s += " (global)"
	}
	return s
}

func AddIgnoreFiles(fs billy.Filesystem, p string) ([]IgnoreFile, error) {
	files, err := fs.ReadDir(p)
	if err != nil {
		return nil, err
	}

	var ifs []IgnoreFile
	for _, f := range files {
		// Check if this is a directory
		if f.IsDir() {
			subFiles, err := AddIgnoreFiles(fs, path.Join(p, f.Name()))
			if err != nil {
				return nil, err
			}
			ifs = append(ifs, subFiles...)
			continue
		}

		// Get the file info
		n := f.Name()
		ext := path.Ext(n)

		// Confirm this is a .gitignore file
		if ext != GitignoreExtension {
			continue
		}

		// Get the filename without the extension
		name := strings.TrimSuffix(n, ext)
		fn := path.Join(p, n)

		// Open the file
		o, err := fs.Open(fn)
		if err != nil {
			return nil, err
		}

		// Read the file
		data, err := io.ReadAll(o)
		o.Close()
		if err != nil {
			return nil, err
		}

		i := IgnoreFile{
			Name: name,
			Path: fn,
			Data: string(data),
		}
		ifs = append(ifs, i)
	}

	return ifs, nil
}

func GetGitignores() ([]IgnoreFile, error) {
	// Create an in-memory file system
	fs := memfs.New()

	// Clone the repository into the memory file system
	_, err := git.Clone(
		memory.NewStorage(),
		fs,
		&git.CloneOptions{
			URL:   RepoURL,
			Depth: 1,
		},
	)
	if err != nil {
		return nil, err
	}

	// Recurse into the file system to get the ignore files
	ifs, err := AddIgnoreFiles(fs, "/")
	if err != nil {
		return nil, err
	}

	return ifs, nil
}
