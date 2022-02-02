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

const RepoURL = "https://github.com/github/gitignore"

type IgnoreFile struct {
	Name string
	Path string
	Data string
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

		// TODO: Check if this is a symlink
		// ...

		// Get the file info
		n := f.Name()
		ext := path.Ext(n)

		// Confirm this is a .gitignore file
		if ext != ".gitignore" {
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
