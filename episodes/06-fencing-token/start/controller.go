package fencing

import (
	"fmt"
	"sync"
)

type Assignment struct {
	HolderID   string
	Generation uint64
}

type Controller struct {
	mu          sync.Mutex
	assignments map[string]Assignment
}

func NewController(assignments map[string]Assignment) *Controller {
	stored := make(map[string]Assignment, len(assignments))
	for manifestID, assignment := range assignments {
		stored[manifestID] = assignment
	}
	return &Controller{assignments: stored}
}

func (c *Controller) Current(manifestID string) (Assignment, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	assignment, ok := c.assignments[manifestID]
	return assignment, ok
}

func (c *Controller) Transfer(
	manifestID string,
	expectedHolder string,
	expectedGeneration uint64,
	newHolder string,
) (Assignment, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	current, ok := c.assignments[manifestID]
	if !ok {
		return Assignment{}, false, fmt.Errorf("manifest %s: assignment not found", manifestID)
	}
	if current.HolderID != expectedHolder || current.Generation != expectedGeneration {
		return Assignment{}, false, nil
	}

	next := Assignment{HolderID: newHolder, Generation: current.Generation + 1}
	c.assignments[manifestID] = next
	return next, true, nil
}
