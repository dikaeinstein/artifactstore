package artifactstore

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"

	"github.com/dikaeinstein/artifactstore/pkg/fsys"
)

type Artifact struct {
	FileType           string
	Name               string
	Url                string
	File               fs.File
	Type               string
	RetrievedFromCache bool
}

type Store interface {
	GetArtifact(url string) (artifact *Artifact, ok bool)
	SaveArtifact(url string, artifact *Artifact) error
}

type ArtifactStore struct {
	client      *http.Client
	downloadDir string
	fs          fs.FS
	store       Store
}

func New(client *http.Client, dlDir string, fs fs.FS, store Store) *ArtifactStore {
	return &ArtifactStore{
		client,
		dlDir,
		fs,
		store,
	}
}

func (afs *ArtifactStore) Get(ctx context.Context, prefix, url string) (*Artifact, error) {
	artifact, found := afs.store.GetArtifact(url)
	if !found {
		artifact, err := afs.downloadArtifact(ctx, prefix, url)
		if err != nil {
			return nil, err
		}

		return artifact, nil
	}

	return artifact, nil
}

func (afs *ArtifactStore) downloadArtifact(ctx context.Context, prefix, url string) (*Artifact, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	res, err := afs.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	artifactName := path.Base(url)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download artifact[%s]: %s", artifactName, res.Status)
	}

	downloadPath := filepath.Join(afs.downloadDir, artifactName)
	file, err := fsys.Create(afs.fs, downloadPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	writableFile, ok := file.(io.Writer)
	if !ok {
		return nil, fmt.Errorf("invalid writer: %T", file)
	}

	n, err := io.Copy(writableFile, res.Body)
	if err != nil {
		return nil, err
	}
	if res.ContentLength != -1 && res.ContentLength != n {
		return nil, fmt.Errorf("copied %v bytes; expected %v", n, res.ContentLength)
	}

	artifact := &Artifact{
		Name:               artifactName,
		FileType:           path.Ext(url),
		Url:                url,
		File:               file,
		RetrievedFromCache: false,
		Type:               prefix,
	}

	err = afs.store.SaveArtifact(url, artifact)
	if err != nil {
		return artifact, err
	}
	return artifact, nil
}
