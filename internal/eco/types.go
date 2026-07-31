package eco

import "time"

const (
	BuildID   = "ECO-V25-20260730-N1-P3"
	BuildName = "Evidence & Casework One Version 25 N1 — Native Evidence & Vision Foundation Preview 3"
	Schema    = 1
)

type Workspace struct {
	Schema       int              `json:"schema"`
	BuildID      string           `json:"build_id"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	Evidence     []EvidenceItem   `json:"evidence"`
	Matters      []Matter         `json:"matters"`
	Changes      []ChangeRecord   `json:"changes"`
	Questions    []QuestionRecord `json:"questions"`
	SelectedID   string           `json:"selected_id,omitempty"`
	SelectedPage string           `json:"selected_page,omitempty"`
	Settings     Settings         `json:"settings"`
}

type Settings struct {
	LowSensory    bool `json:"low_sensory"`
	ReducedMotion bool `json:"reduced_motion"`
}

type EvidenceItem struct {
	ID              string           `json:"id"`
	OriginalName    string           `json:"original_name"`
	SafeName        string           `json:"safe_name"`
	SourcePath      string           `json:"source_path,omitempty"`
	Size            int64            `json:"size"`
	SHA256          string           `json:"sha256"`
	DetectedType    string           `json:"detected_type"`
	ExtensionType   string           `json:"extension_type,omitempty"`
	TypeMismatch    bool             `json:"type_mismatch"`
	Readable        bool             `json:"readable"`
	Status          string           `json:"status"`
	ImportedAt      time.Time        `json:"imported_at"`
	ObjectFile      string           `json:"object_file"`
	ExtractedText   string           `json:"extracted_text,omitempty"`
	Segments        []SourceSegment  `json:"segments,omitempty"`
	Image           *ImageAssessment `json:"image,omitempty"`
	Rotation        int              `json:"rotation,omitempty"`
	DuplicateOf     string           `json:"duplicate_of,omitempty"`
	NearDuplicateOf string           `json:"near_duplicate_of,omitempty"`
	Warnings        []string         `json:"warnings,omitempty"`
	MatterIDs       []string         `json:"matter_ids,omitempty"`
}

type ImageAssessment struct {
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	Megapixels     float64  `json:"megapixels"`
	Brightness     float64  `json:"brightness"`
	Contrast       float64  `json:"contrast"`
	BlurVariance   float64  `json:"blur_variance"`
	EdgeDensity    float64  `json:"edge_density"`
	PerceptualHash string   `json:"perceptual_hash,omitempty"`
	Orientation    string   `json:"orientation"`
	QualityLabel   string   `json:"quality_label"`
	Warnings       []string `json:"warnings,omitempty"`
}

type SourceSegment struct {
	ID       string `json:"id"`
	Ordinal  int    `json:"ordinal"`
	Text     string `json:"text"`
	PageHint string `json:"page_hint,omitempty"`
}

type Matter struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Reference   string    `json:"reference,omitempty"`
	Objective   string    `json:"objective,omitempty"`
	Status      string    `json:"status"`
	NextAction  string    `json:"next_action,omitempty"`
	EvidenceIDs []string  `json:"evidence_ids,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ChangeRecord struct {
	ID       string         `json:"id"`
	At       time.Time      `json:"at"`
	Actor    string         `json:"actor"`
	Type     string         `json:"type"`
	Summary  string         `json:"summary"`
	Details  map[string]any `json:"details,omitempty"`
	PrevHash string         `json:"prev_hash,omitempty"`
	Hash     string         `json:"hash"`
}

type QuestionRecord struct {
	ID                        string     `json:"id"`
	AskedAt                   time.Time  `json:"asked_at"`
	Question                  string     `json:"question"`
	Intent                    string     `json:"intent"`
	Answer                    string     `json:"answer"`
	Citations                 []Citation `json:"citations,omitempty"`
	Support                   string     `json:"support"`
	ScopeIDs                  []string   `json:"scope_ids,omitempty"`
	ReceiptID                 string     `json:"receipt_id"`
	EvidenceConsidered        int        `json:"evidence_considered"`
	RetrievedSegments         int        `json:"retrieved_segments"`
	SuspiciousSourcesExcluded int        `json:"suspicious_sources_excluded"`
}

type Citation struct {
	EvidenceID string  `json:"evidence_id"`
	SegmentID  string  `json:"segment_id"`
	Label      string  `json:"label"`
	Quote      string  `json:"quote"`
	Score      float64 `json:"score"`
}

type ImportProgress struct {
	Path    string
	Name    string
	Stage   string
	Current int64
	Total   int64
}
