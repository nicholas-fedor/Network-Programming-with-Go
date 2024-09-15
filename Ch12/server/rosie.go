// Pages 286-287
// Listing 12-21: Building a RobotMaid-compatible type named Rosie.
package main

import (
	"context"
	"fmt"
	"sync"

	"Ch12/housework/v1"
)

type Rosie struct {
	mu sync.Mutex
	// The new Rosie struct keeps its list of chores in memory, guarded by a
	// mutex, since more than one client can concurrently use the service.
	chores []*housework.Chore
}

func (r *Rosie) Add(_ context.Context, chores *housework.Chores) (*housework.Response, error) {
	r.mu.Lock()
	r.chores = append(r.chores, chores.Chores...)
	r.mu.Unlock()

	// The Add, Complete, and List methods all return either a response message
	// type or an error, both of which ultimately make their way back to the client.
	return &housework.Response{Message: "ok"}, nil
}

func (r *Rosie) Complete(_ context.Context, req *housework.CompleteRequest) (*housework.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.chores == nil || req.ChoreNumber < 1 || int(req.ChoreNumber) > len(r.chores) {
		return nil, fmt.Errorf("chore %d not found", req.ChoreNumber)
	}

	r.chores[req.ChoreNumber].Complete = true

	return &housework.Response{Message: "ok"}, nil
}

func (r *Rosie) List(_ context.Context, _ *housework.Empty) (*housework.Chores, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.chores == nil {
		r.chores = make([]*housework.Chore, 0)
	}

	return &housework.Chores{Chores: r.chores}, nil
}

func (r *Rosie) Service() *housework.RobotMaidService {
	// The Service method returns a pointer to a new housework.RobotMaidService
	// instance, where Rosie's Add, Complete, and List methods map their
	// corresponding method on the new instance.
	return &housework.RobotMaidService{
		Add:      r.Add,
		Complete: r.Complete,
		List:     r.List,
	}
}
