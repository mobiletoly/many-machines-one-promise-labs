package fencing

import (
	"fmt"
	"sync"
)

type PublishResult string

const (
	PublicationAccepted        PublishResult = "accepted"
	PublicationStaleGeneration PublishResult = "stale_generation"
)

type Manifest struct {
	ID                  string
	HighestGeneration   uint64
	OfficialContent     string
	PublishedBy         string
	PublishedGeneration uint64
}

type PublicationStore struct {
	mu        sync.Mutex
	manifests map[string]Manifest
}

func NewPublicationStore(manifests ...Manifest) *PublicationStore {
	stored := make(map[string]Manifest, len(manifests))
	for _, manifest := range manifests {
		stored[manifest.ID] = manifest
	}
	return &PublicationStore{manifests: stored}
}

func (s *PublicationStore) Manifest(id string) (Manifest, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	manifest, ok := s.manifests[id]
	return manifest, ok
}

func (s *PublicationStore) Publish(
	manifestID string,
	holderID string,
	generation uint64,
	content string,
) (PublishResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	manifest, ok := s.manifests[manifestID]
	if !ok {
		return "", fmt.Errorf("manifest %s: not found", manifestID)
	}

	if generation < manifest.HighestGeneration {
		return PublicationStaleGeneration, nil
	}
	if generation > manifest.HighestGeneration {
		manifest.HighestGeneration = generation
	}
	manifest.OfficialContent = content
	manifest.PublishedBy = holderID
	manifest.PublishedGeneration = generation
	s.manifests[manifestID] = manifest

	return PublicationAccepted, nil
}
