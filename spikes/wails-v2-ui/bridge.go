package main

import (
	"context"

	application "github.com/ECO-evidence-casework-one/eco/internal/app"
)

// Bridge is intentionally transport-only. Business and workspace safety logic
// stays in internal/app and internal/eco.
type Bridge struct {
	ctx     context.Context
	service *application.Service
}

func NewBridge() *Bridge {
	return &Bridge{service: &application.Service{}}
}

func (b *Bridge) startup(ctx context.Context) {
	b.ctx = ctx
}

func (b *Bridge) shutdown(context.Context) {
	_ = b.service.CloseWorkspace()
}

func (b *Bridge) OpenWorkspace(root string) error {
	return b.service.OpenWorkspace(root)
}

func (b *Bridge) CreateWorkspace(root string) error {
	return b.service.CreateWorkspace(root)
}

func (b *Bridge) CloseWorkspace() error {
	return b.service.CloseWorkspace()
}

func (b *Bridge) SnapshotSummary() (application.WorkspaceSummary, error) {
	return b.service.SnapshotSummary()
}

func (b *Bridge) CreateMatter(title string) error {
	return b.service.CreateMatter(title)
}
