package eco

import "time"

const (
	BuildID   = "ECO-V25-20260731-N2-P1"
	BuildName = "Evidence & Casework One Version 25 N2 — Native Document Vision Foundation Preview 1"
	Schema    = 1
)

type Workspace struct {
	Schema        int                  `json:"schema"`
	BuildID       string               `json:"build_id"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	Evidence      []EvidenceItem       `json:"evidence"`
	Preservations []PreservationRecord `json:"preservations,omitempty"`
	Matters       []Matter             `json:"matters"`
	Changes       []ChangeRecord       `json:"changes"`
	Questions     []QuestionRecord     `json:"questions"`
	SelectedID    string               `json:"selected_id,omitempty"`
	SelectedPage  string               `json:"selected_page,omitempty"`
	Settings      Settings             `json:"settings"`
}

type Settings struct {
	LowSensory    bool `json:"low_sensory"`
	ReducedMotion bool `json:"reduced_motion"`
}

type EvidenceItem struct {
	ID                string             `json:"id"`
	OriginalName      string             `json:"original_name"`
	SafeName          string             `json:"safe_name"`
	SourcePath        string             `json:"source_path,omitempty"`
	Size              int64              `json:"size"`
	SHA256            string             `json:"sha256"`
	DetectedType      string             `json:"detected_type"`
	ExtensionType     string             `json:"extension_type,omitempty"`
	TypeMismatch      bool               `json:"type_mismatch"`
	Readable          bool               `json:"readable"`
	Status            string             `json:"status"`
	ImportedAt        time.Time          `json:"imported_at"`
	ObjectFile        string             `json:"object_file"`
	Preservation      string             `json:"preservation"`
	SourceVerified    bool               `json:"source_verified"`
	SourceVerifiedAt  time.Time          `json:"source_verified_at,omitempty"`
	VerificationError string             `json:"verification_error,omitempty"`
	ExtractedText     string             `json:"extracted_text,omitempty"`
	Extraction        *ExtractionReceipt `json:"extraction,omitempty"`
	Segments          []SourceSegment    `json:"segments,omitempty"`
	Image             *ImageAssessment   `json:"image,omitempty"`
	OCR               *OCRReceipt        `json:"ocr,omitempty"`
	Rotation          int                `json:"rotation,omitempty"`
	DuplicateOf       string             `json:"duplicate_of,omitempty"`
	NearDuplicateOf   string             `json:"near_duplicate_of,omitempty"`
	Warnings          []string           `json:"warnings,omitempty"`
	MatterIDs         []string           `json:"matter_ids,omitempty"`
}

// PreservationRecord is the durable state machine for an import. It is kept
// outside Evidence so an interrupted or failed write can never look like a
// completed, indexable evidence item.
type PreservationRecord struct {
	ID              string    `json:"id"`
	EvidenceID      string    `json:"evidence_id"`
	ObjectFile      string    `json:"object_file"`
	OriginalName    string    `json:"original_name"`
	SafeName        string    `json:"safe_name"`
	SourcePath      string    `json:"source_path,omitempty"`
	State           string    `json:"state"`
	ExpectedSize    int64     `json:"expected_size"`
	BytesPreserved  int64     `json:"bytes_preserved,omitempty"`
	IntakeSHA256    string    `json:"intake_sha256,omitempty"`
	PreservedSHA256 string    `json:"preserved_sha256,omitempty"`
	StartedAt       time.Time `json:"started_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	VerifiedAt      time.Time `json:"verified_at,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// SourceReceipt binds a downstream operation to the immutable encrypted
// object and to a fresh verification of its decrypted bytes.
type SourceReceipt struct {
	EvidenceID string    `json:"evidence_id"`
	ObjectFile string    `json:"object_file"`
	SHA256     string    `json:"sha256"`
	Size       int64     `json:"size"`
	VerifiedAt time.Time `json:"verified_at"`
}

type ExtractionReceipt struct {
	SourceObject string    `json:"source_object"`
	SourceSHA256 string    `json:"source_sha256"`
	DetectedType string    `json:"detected_type"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	Segments     int       `json:"segments"`
}

type ImageAssessment struct {
	SourceObject          string          `json:"source_object"`
	SourceSHA256          string          `json:"source_sha256"`
	Width                 int             `json:"width"`
	Height                int             `json:"height"`
	Megapixels            float64         `json:"megapixels"`
	Brightness            float64         `json:"brightness"`
	Contrast              float64         `json:"contrast"`
	BlurVariance          float64         `json:"blur_variance"`
	EdgeDensity           float64         `json:"edge_density"`
	PerceptualHash        string          `json:"perceptual_hash,omitempty"`
	Orientation           string          `json:"orientation"`
	QualityLabel          string          `json:"quality_label"`
	SkewCorrectionDegrees float64         `json:"skew_correction_degrees,omitempty"`
	SkewConfidence        float64         `json:"skew_confidence,omitempty"`
	GlareRatio            float64         `json:"glare_ratio,omitempty"`
	ShadowImbalance       float64         `json:"shadow_imbalance,omitempty"`
	BorderInkRatio        float64         `json:"border_ink_ratio,omitempty"`
	ProbableDoublePage    bool            `json:"probable_double_page,omitempty"`
	SuggestedCrop         *CropSuggestion `json:"suggested_crop,omitempty"`
	Warnings              []string        `json:"warnings,omitempty"`
}

type NormalizedRegion struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

func (r NormalizedRegion) Valid() bool {
	return r.X >= 0 && r.Y >= 0 && r.Width > 0 && r.Height > 0 && r.X+r.Width <= 1.000001 && r.Y+r.Height <= 1.000001
}

type OCRWord struct {
	Text       string           `json:"text"`
	Confidence float64          `json:"confidence"`
	Region     NormalizedRegion `json:"region"`
	Page       int              `json:"page"`
}

type OCRLine struct {
	Text       string           `json:"text"`
	Confidence float64          `json:"confidence"`
	Region     NormalizedRegion `json:"region"`
	Page       int              `json:"page"`
	Words      []OCRWord        `json:"words,omitempty"`
}

type OCRReceipt struct {
	Engine            string    `json:"engine"`
	EngineVersion     string    `json:"engine_version,omitempty"`
	Language          string    `json:"language,omitempty"`
	Status            string    `json:"status"`
	SourceObject      string    `json:"source_object"`
	SourceSHA256      string    `json:"source_sha256"`
	CreatedAt         time.Time `json:"created_at"`
	AverageConfidence float64   `json:"average_confidence,omitempty"`
	Words             []OCRWord `json:"words,omitempty"`
	Lines             []OCRLine `json:"lines,omitempty"`
	Warnings          []string  `json:"warnings,omitempty"`
}

type SourceSegment struct {
	ID           string            `json:"id"`
	Ordinal      int               `json:"ordinal"`
	Text         string            `json:"text"`
	PageHint     string            `json:"page_hint,omitempty"`
	Page         int               `json:"page,omitempty"`
	Region       *NormalizedRegion `json:"region,omitempty"`
	Origin       string            `json:"origin,omitempty"`
	Confidence   float64           `json:"confidence,omitempty"`
	SourceObject string            `json:"source_object"`
	SourceSHA256 string            `json:"source_sha256"`
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
	ID                           string     `json:"id"`
	AskedAt                      time.Time  `json:"asked_at"`
	Question                     string     `json:"question"`
	Intent                       string     `json:"intent"`
	Answer                       string     `json:"answer"`
	Citations                    []Citation `json:"citations,omitempty"`
	Support                      string     `json:"support"`
	ScopeIDs                     []string   `json:"scope_ids,omitempty"`
	ReceiptID                    string     `json:"receipt_id"`
	EvidenceConsidered           int        `json:"evidence_considered"`
	RetrievedSegments            int        `json:"retrieved_segments"`
	SuspiciousSourcesExcluded    int        `json:"suspicious_sources_excluded"`
	LowConfidenceSourcesExcluded int        `json:"low_confidence_sources_excluded"`
	SourceVerificationFailures   int        `json:"source_verification_failures"`
}

type Citation struct {
	EvidenceID   string            `json:"evidence_id"`
	SegmentID    string            `json:"segment_id"`
	Label        string            `json:"label"`
	Quote        string            `json:"quote"`
	Score        float64           `json:"score"`
	Page         int               `json:"page,omitempty"`
	Region       *NormalizedRegion `json:"region,omitempty"`
	Origin       string            `json:"origin,omitempty"`
	Confidence   float64           `json:"confidence,omitempty"`
	SourceObject string            `json:"source_object"`
	SourceSHA256 string            `json:"source_sha256"`
}

type ImportProgress struct {
	Path    string
	Name    string
	Stage   string
	Current int64
	Total   int64
}
