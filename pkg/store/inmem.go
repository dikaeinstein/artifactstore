package store

import "github.com/dikaeinstein/artifactstore/artifactstore"

type InMemStore struct {
	m map[string]*artifactstore.Artifact
}

func NewInMemStore(m map[string]*artifactstore.Artifact) *InMemStore {
	return &InMemStore{m}
}

func (inmem *InMemStore) GetArtifact(url string) (*artifactstore.Artifact, bool) {
	artifact, ok := inmem.m[url]
	if ok {
		artifact.RetrievedFromCache = true
	}

	return artifact, ok
}

func (inmem *InMemStore) SaveArtifact(url string, artifact *artifactstore.Artifact) error {
	inmem.m[url] = artifact
	return nil
}
