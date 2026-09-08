package app

import (
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/ECO-evidence-casework-one/eco/internal/eco"
)

var (
	ErrWorkspaceAlreadyOpen = errors.New("a workspace is already open")
	ErrNoWorkspaceOpen      = errors.New("no workspace is open")
)

// Service is ECO's UI-neutral application boundary.
//
// UI transports should depend on this layer instead of receiving the Vault,
// encryption material, object filenames, source paths, or the complete
// persisted Workspace schema.
type Service struct {
	mu    sync.Mutex
	vault *eco.Vault
}

type WorkspaceSummary struct {
	Root              string          `json:"root"`
	Schema            int             `json:"schema"`
	Revision          uint64          `json:"revision"`
	BuildID           string          `json:"build_id"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	SelectedPage      string          `json:"selected_page,omitempty"`
	MatterCount       int             `json:"matter_count"`
	EvidenceCount     int             `json:"evidence_count"`
	PreservationCount int             `json:"preservation_count"`
	Matters           []MatterSummary `json:"matters"`
}

type MatterSummary struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Reference     string    `json:"reference,omitempty"`
	Objective     string    `json:"objective,omitempty"`
	Status        string    `json:"status"`
	NextAction    string    `json:"next_action,omitempty"`
	EvidenceCount int       `json:"evidence_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (s *Service) OpenWorkspace(root string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vault != nil {
		return ErrWorkspaceAlreadyOpen
	}
	v, err := eco.OpenVault(root)
	if err != nil {
		return err
	}
	s.vault = v
	return nil
}

func (s *Service) CreateWorkspace(root string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vault != nil {
		return ErrWorkspaceAlreadyOpen
	}
	v, err := eco.CreateVault(root)
	if err != nil {
		return err
	}
	s.vault = v
	return nil
}

func (s *Service) CloseWorkspace() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vault == nil {
		return nil
	}
	v := s.vault
	s.vault = nil
	return v.Close()
}

func (s *Service) SnapshotSummary() (WorkspaceSummary, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vault == nil {
		return WorkspaceSummary{}, ErrNoWorkspaceOpen
	}
	return summarize(filepath.Clean(s.vault.Root), s.vault.Snapshot()), nil
}

func (s *Service) CreateMatter(title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.vault == nil {
		return ErrNoWorkspaceOpen
	}
	_, err := s.vault.CreateMatter(title)
	return err
}

func summarize(root string, ws eco.Workspace) WorkspaceSummary {
	out := WorkspaceSummary{
		Root:              root,
		Schema:            ws.Schema,
		Revision:          ws.Revision,
		BuildID:           ws.BuildID,
		CreatedAt:         ws.CreatedAt,
		UpdatedAt:         ws.UpdatedAt,
		SelectedPage:      ws.SelectedPage,
		MatterCount:       len(ws.Matters),
		EvidenceCount:     len(ws.Evidence),
		PreservationCount: len(ws.Preservations),
		Matters:           make([]MatterSummary, 0, len(ws.Matters)),
	}
	for _, matter := range ws.Matters {
		out.Matters = append(out.Matters, MatterSummary{
			ID:            matter.ID,
			Title:         matter.Title,
			Reference:     matter.Reference,
			Objective:     matter.Objective,
			Status:        matter.Status,
			NextAction:    matter.NextAction,
			EvidenceCount: len(matter.EvidenceIDs),
			CreatedAt:     matter.CreatedAt,
			UpdatedAt:     matter.UpdatedAt,
		})
	}
	return out
}
