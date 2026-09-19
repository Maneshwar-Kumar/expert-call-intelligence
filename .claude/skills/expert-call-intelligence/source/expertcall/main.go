package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	toolVersion    = "1.0.0"
	outputRootName = "_Expert_Call_System"
)

type Config struct {
	SchemaVersion      string `json:"schema_version"`
	ToolVersion        string `json:"tool_version"`
	RawInputRule       string `json:"raw_input_rule"`
	OutputRoot         string `json:"output_root"`
	MasterFilename     string `json:"master_filename"`
	CallIDPrefix       string `json:"call_id_prefix"`
	ProcessRootOnly    bool   `json:"process_root_only"`
	MaxCompileJobItems int    `json:"max_compile_job_items"`
}

type Registry struct {
	SchemaVersion string      `json:"schema_version"`
	NextSequence  int         `json:"next_sequence"`
	Calls         []CallEntry `json:"calls"`
}

type CallEntry struct {
	CallID              string `json:"call_id"`
	Sequence            int    `json:"sequence"`
	SourceFile          string `json:"source_file"`
	SourceSHA256        string `json:"source_sha256"`
	SourceSize          int64  `json:"source_size"`
	SourceModifiedUTC   string `json:"source_modified_utc"`
	Status              string `json:"status"`
	FirstSeenUTC        string `json:"first_seen_utc"`
	LastProcessedUTC    string `json:"last_processed_utc,omitempty"`
	ExtractedPath       string `json:"extracted_path,omitempty"`
	SummarySpecPath     string `json:"summary_spec_path,omitempty"`
	CallRecordPath      string `json:"call_record_path,omitempty"`
	SummaryDocxPath     string `json:"summary_docx_path,omitempty"`
	SummaryMarkdownPath string `json:"summary_markdown_path,omitempty"`
	Error               string `json:"error,omitempty"`
	SourceMissing       bool   `json:"source_missing,omitempty"`
}

type SummarySpec struct {
	SchemaVersion string      `json:"schema_version"`
	CallID        string      `json:"call_id"`
	SourceFile    string      `json:"source_file"`
	SourceSHA256  string      `json:"source_sha256"`
	Interviews    []Interview `json:"interviews"`
}

type Interview struct {
	Title   string   `json:"title"`
	Framing []string `json:"framing"`
	Themes  []Theme  `json:"themes"`
}

type Theme struct {
	Lead   string   `json:"lead"`
	Points []string `json:"points"`
}

type Person struct {
	Name    string `json:"name,omitempty"`
	Role    string `json:"role,omitempty"`
	Company string `json:"company,omitempty"`
}

type Entity struct {
	CanonicalName string   `json:"canonical_name"`
	DisplayName   string   `json:"display_name,omitempty"`
	EntityType    string   `json:"entity_type,omitempty"`
	Roles         []string `json:"roles,omitempty"`
	Aliases       []string `json:"aliases,omitempty"`
}

type Metric struct {
	Name        string      `json:"name,omitempty"`
	Value       interface{} `json:"value,omitempty"`
	RangeLow    interface{} `json:"range_low,omitempty"`
	RangeHigh   interface{} `json:"range_high,omitempty"`
	Unit        string      `json:"unit,omitempty"`
	Currency    string      `json:"currency,omitempty"`
	Denominator string      `json:"denominator,omitempty"`
	Period      string      `json:"period,omitempty"`
	Scenario    string      `json:"scenario,omitempty"`
}

type Evidence struct {
	EvidenceID    string                 `json:"evidence_id"`
	Assertion     string                 `json:"assertion"`
	Subject       string                 `json:"subject,omitempty"`
	Predicate     string                 `json:"predicate,omitempty"`
	Object        string                 `json:"object,omitempty"`
	Modules       []string               `json:"modules"`
	Topic         string                 `json:"topic"`
	StatementType string                 `json:"statement_type"`
	Basis         string                 `json:"basis"`
	Confidence    string                 `json:"confidence"`
	TimeScope     map[string]interface{} `json:"time_scope,omitempty"`
	Geography     []string               `json:"geography,omitempty"`
	ProductScope  []string               `json:"product_scope,omitempty"`
	CustomerScope []string               `json:"customer_scope,omitempty"`
	Scenario      string                 `json:"scenario,omitempty"`
	Metric        *Metric                `json:"metric,omitempty"`
	SourceAnchors []string               `json:"source_anchors"`
	Caveats       []string               `json:"caveats,omitempty"`
	Keywords      []string               `json:"keywords,omitempty"`
}

type Relationship struct {
	From        string                 `json:"from"`
	Type        string                 `json:"type"`
	To          string                 `json:"to"`
	Scope       map[string]interface{} `json:"scope,omitempty"`
	EvidenceIDs []string               `json:"evidence_ids,omitempty"`
	Confidence  string                 `json:"confidence,omitempty"`
}

type OpenQuestion struct {
	Question          string   `json:"question"`
	WhyItMatters      string   `json:"why_it_matters,omitempty"`
	Priority          string   `json:"priority,omitempty"`
	RelatedEvidenceID []string `json:"related_evidence_ids,omitempty"`
}

type CallRecord struct {
	SchemaVersion   string         `json:"schema_version"`
	CallID          string         `json:"call_id"`
	SourceFile      string         `json:"source_file"`
	SourceSHA256    string         `json:"source_sha256"`
	InterviewTitle  string         `json:"interview_title"`
	Interviewee     Person         `json:"interviewee"`
	StakeholderType string         `json:"stakeholder_type"`
	Focus           string         `json:"focus"`
	ScopeNotes      []string       `json:"scope_notes,omitempty"`
	Entities        []Entity       `json:"entities,omitempty"`
	Evidence        []Evidence     `json:"evidence"`
	Relationships   []Relationship `json:"relationships,omitempty"`
	OpenQuestions   []OpenQuestion `json:"open_questions,omitempty"`
	Keywords        []string       `json:"keywords,omitempty"`
}

type SourceRef struct {
	CallID              string `json:"call_id"`
	Sequence            int    `json:"sequence"`
	SourceFile          string `json:"source_file"`
	SourceSHA256        string `json:"source_sha256"`
	InterviewTitle      string `json:"interview_title"`
	StakeholderType     string `json:"stakeholder_type"`
	Focus               string `json:"focus"`
	ExpertName          string `json:"expert_name,omitempty"`
	ExpertRole          string `json:"expert_role,omitempty"`
	ExpertCompany       string `json:"expert_company,omitempty"`
	ExtractedPath       string `json:"extracted_path"`
	SummaryDocxPath     string `json:"summary_docx_path"`
	SummaryMarkdownPath string `json:"summary_markdown_path"`
}

type EvidenceLedgerRow struct {
	Evidence Evidence  `json:"evidence"`
	Source   SourceRef `json:"source"`
}

type JobGroup struct {
	Module        string `json:"module"`
	Topic         string `json:"topic"`
	PrimaryEntity string `json:"primary_entity"`
}

type CompileJob struct {
	SchemaVersion string              `json:"schema_version"`
	JobID         string              `json:"job_id"`
	InputHash     string              `json:"input_hash"`
	Group         JobGroup            `json:"group"`
	OutputPath    string              `json:"output_path"`
	Evidence      []EvidenceLedgerRow `json:"evidence"`
}

type CompileJobManifestEntry struct {
	JobID      string   `json:"job_id"`
	InputHash  string   `json:"input_hash"`
	Group      JobGroup `json:"group"`
	JobPath    string   `json:"job_path"`
	OutputPath string   `json:"output_path"`
	Status     string   `json:"status"`
	ItemCount  int      `json:"item_count"`
}

type CompileJobManifest struct {
	SchemaVersion string                    `json:"schema_version"`
	GeneratedUTC  string                    `json:"generated_utc"`
	Jobs          []CompileJobManifestEntry `json:"jobs"`
}

type Claim struct {
	ClaimKey              string                 `json:"claim_key"`
	NormalizedClaim       string                 `json:"normalized_claim"`
	Synthesis             string                 `json:"synthesis"`
	Status                string                 `json:"status"`
	Confidence            string                 `json:"confidence"`
	PrimaryEntities       []string               `json:"primary_entities,omitempty"`
	Scope                 map[string]interface{} `json:"scope,omitempty"`
	SupportingEvidence    []string               `json:"supporting_evidence,omitempty"`
	ContradictingEvidence []string               `json:"contradicting_evidence,omitempty"`
	QualifyingEvidence    []string               `json:"qualifying_evidence,omitempty"`
	OpenQuestions         []string               `json:"open_questions,omitempty"`
}

type ClaimShard struct {
	SchemaVersion       string   `json:"schema_version"`
	JobID               string   `json:"job_id"`
	InputHash           string   `json:"input_hash"`
	Group               JobGroup `json:"group"`
	Claims              []Claim  `json:"claims"`
	UnclusteredEvidence []string `json:"unclustered_evidence,omitempty"`
}

type FinalClaim struct {
	ClaimID               string                 `json:"claim_id"`
	ClaimKey              string                 `json:"claim_key"`
	NormalizedClaim       string                 `json:"normalized_claim"`
	Synthesis             string                 `json:"synthesis"`
	Status                string                 `json:"status"`
	Confidence            string                 `json:"confidence"`
	Module                string                 `json:"module"`
	Topic                 string                 `json:"topic"`
	PrimaryEntity         string                 `json:"primary_entity"`
	PrimaryEntities       []string               `json:"primary_entities,omitempty"`
	Scope                 map[string]interface{} `json:"scope,omitempty"`
	SupportingEvidence    []string               `json:"supporting_evidence,omitempty"`
	ContradictingEvidence []string               `json:"contradicting_evidence,omitempty"`
	QualifyingEvidence    []string               `json:"qualifying_evidence,omitempty"`
	OpenQuestions         []string               `json:"open_questions,omitempty"`
}

type EntityAggregate struct {
	CanonicalName string         `json:"canonical_name"`
	DisplayName   string         `json:"display_name"`
	EntityType    string         `json:"entity_type,omitempty"`
	Roles         []string       `json:"roles,omitempty"`
	Aliases       []string       `json:"aliases,omitempty"`
	SourceCalls   []string       `json:"source_calls,omitempty"`
	EvidenceCount int            `json:"evidence_count"`
	ClaimCount    int            `json:"claim_count"`
	Modules       map[string]int `json:"modules,omitempty"`
	Topics        map[string]int `json:"topics,omitempty"`
}

type DiscoverOutput struct {
	Pending    []CallEntry        `json:"pending"`
	Unchanged  []CallEntry        `json:"unchanged"`
	Missing    []CallEntry        `json:"missing"`
	Duplicates []DuplicateFinding `json:"duplicates"`
	Summary    map[string]int     `json:"summary"`
}

type DuplicateFinding struct {
	SourceFile  string `json:"source_file"`
	DuplicateOf string `json:"duplicate_of"`
	SHA256      string `json:"sha256"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	var err error
	switch cmd {
	case "version":
		fmt.Println(toolVersion)
		return
	case "init":
		err = cmdInit(os.Args[2:])
	case "discover":
		err = cmdDiscover(os.Args[2:])
	case "extract":
		err = cmdExtract(os.Args[2:])
	case "validate":
		err = cmdValidate(os.Args[2:])
	case "render-summary":
		err = cmdRenderSummary(os.Args[2:])
	case "mark-complete":
		err = cmdMarkComplete(os.Args[2:])
	case "mark-failed":
		err = cmdMarkFailed(os.Args[2:])
	case "build-ledger":
		err = cmdBuildLedger(os.Args[2:])
	case "discover-jobs":
		err = cmdDiscoverJobs(os.Args[2:])
	case "validate-shard":
		err = cmdValidateShard(os.Args[2:])
	case "finalize-index":
		err = cmdFinalizeIndex(os.Args[2:])
	case "render-master":
		err = cmdRenderMaster(os.Args[2:])
	case "search":
		err = cmdSearch(os.Args[2:])
	case "trace":
		err = cmdTrace(os.Args[2:])
	case "status":
		err = cmdStatus(os.Args[2:])
	case "self-test":
		err = cmdSelfTest(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `expertcall - deterministic file and index utility for the Expert Call Intelligence plugin

Commands:
  init              Create output folders and configuration
  discover          Register and list new or changed root-level DOCX transcripts
  extract           Extract one DOCX to anchored Markdown
  validate          Validate one summary spec and call record
  render-summary    Render one standalone summary DOCX and Markdown
  mark-complete     Mark a processed call complete in the registry
  mark-failed       Record a processing failure
  build-ledger      Rebuild evidence ledger and create semantic compile jobs
  discover-jobs     List pending semantic compile jobs
  validate-shard    Validate one claim shard against its compile job
  finalize-index    Merge claim shards and build case maps/indexes
  render-master     Rebuild the consolidated Word compilation
  search            Retrieve a compact query packet from the case pack
  trace             Retrieve exact source paragraphs for evidence or claims
  status            Show processing and compilation status
  self-test         Run deterministic internal tests
  version           Print version`)
}

func rootFlag(args []string, name string) (*flag.FlagSet, *string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	root := fs.String("root", ".", "case folder root")
	return fs, root
}

func nowUTC() string { return time.Now().UTC().Format(time.RFC3339) }

func absRoot(root string) (string, error) {
	a, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.Clean(a), nil
}

func paths(root string) map[string]string {
	out := filepath.Join(root, outputRootName)
	sys := filepath.Join(out, "_System")
	return map[string]string{
		"out":       out,
		"processed": filepath.Join(out, "01_Processed_Summaries"),
		"knowledge": filepath.Join(out, "02_Knowledge_Pack"),
		"compiled":  filepath.Join(out, "03_Compiled_Output"),
		"review":    filepath.Join(out, "04_Review_Queue"),
		"system":    sys,
		"extracted": filepath.Join(sys, "extracted"),
		"specs":     filepath.Join(sys, "specs"),
		"records":   filepath.Join(sys, "records"),
		"jobs":      filepath.Join(sys, "compile_jobs"),
		"shards":    filepath.Join(sys, "claim_shards"),
		"logs":      filepath.Join(sys, "logs"),
		"query":     filepath.Join(sys, "query_packets"),
		"registry":  filepath.Join(sys, "processing_registry.json"),
		"config":    filepath.Join(sys, "config.json"),
		"manifest":  filepath.Join(sys, "compile_jobs_manifest.json"),
	}
}

func defaultConfig() Config {
	return Config{
		SchemaVersion:      "1.0",
		ToolVersion:        toolVersion,
		RawInputRule:       "Process root-level .docx files only; never modify source transcripts",
		OutputRoot:         outputRootName,
		MasterFilename:     "Expert_Call_Notes_Master.docx",
		CallIDPrefix:       "EC",
		ProcessRootOnly:    true,
		MaxCompileJobItems: 120,
	}
}

func ensureDirs(root string) error {
	p := paths(root)
	dirs := []string{p["out"], p["processed"], p["knowledge"], p["compiled"], p["review"], p["system"], p["extracted"], p["specs"], p["records"], p["jobs"], p["shards"], p["logs"], p["query"], filepath.Join(p["knowledge"], "Entities"), filepath.Join(p["knowledge"], "Sources")}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	if _, err := os.Stat(p["config"]); errors.Is(err, os.ErrNotExist) {
		if err := writeJSONAtomic(p["config"], defaultConfig()); err != nil {
			return err
		}
	}
	if _, err := os.Stat(p["registry"]); errors.Is(err, os.ErrNotExist) {
		r := Registry{SchemaVersion: "1.0", NextSequence: 1, Calls: []CallEntry{}}
		if err := writeJSONAtomic(p["registry"], r); err != nil {
			return err
		}
	}
	return nil
}

func loadConfig(root string) (Config, error) {
	if err := ensureDirs(root); err != nil {
		return Config{}, err
	}
	var c Config
	if err := readJSON(paths(root)["config"], &c); err != nil {
		return Config{}, err
	}
	if c.CallIDPrefix == "" {
		c.CallIDPrefix = "EC"
	}
	if c.MasterFilename == "" {
		c.MasterFilename = "Expert_Call_Notes_Master.docx"
	}
	if c.MaxCompileJobItems <= 0 {
		c.MaxCompileJobItems = 120
	}
	return c, nil
}

func loadRegistry(root string) (Registry, error) {
	if err := ensureDirs(root); err != nil {
		return Registry{}, err
	}
	var r Registry
	if err := readJSON(paths(root)["registry"], &r); err != nil {
		return Registry{}, err
	}
	if r.NextSequence <= 0 {
		maxSeq := 0
		for _, c := range r.Calls {
			if c.Sequence > maxSeq {
				maxSeq = c.Sequence
			}
		}
		r.NextSequence = maxSeq + 1
	}
	return r, nil
}

func saveRegistry(root string, r Registry) error { return writeJSONAtomic(paths(root)["registry"], r) }

func readJSON(path string, v interface{}) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func writeJSONAtomic(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}
	return os.Rename(tmp, path)
}

func writeTextAtomic(path, text string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(text), 0644); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}
	return os.Rename(tmp, path)
}

func relPath(root, p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	return filepath.ToSlash(r)
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashBytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hashString(s string) string { return hashBytes([]byte(s)) }

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteByte('-')
			dash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "general"
	}
	return out
}

func safeFilename(s string) string {
	base := strings.TrimSuffix(filepath.Base(s), filepath.Ext(s))
	var b strings.Builder
	for _, r := range base {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ' ' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := strings.Join(strings.Fields(b.String()), "_")
	if len(out) > 80 {
		out = out[:80]
	}
	if out == "" {
		out = "Expert_Call"
	}
	return out
}

func uniqueSorted(xs []string) []string {
	m := map[string]string{}
	for _, x := range xs {
		t := strings.TrimSpace(x)
		if t == "" {
			continue
		}
		k := strings.ToLower(t)
		if _, ok := m[k]; !ok {
			m[k] = t
		}
	}
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}

func cmdInit(args []string) error {
	fs, rootFlagVal := rootFlag(args, "init")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	if err := ensureDirs(root); err != nil {
		return err
	}
	out := map[string]interface{}{"root": root, "output_root": paths(root)["out"], "tool_version": toolVersion, "status": "initialized"}
	return printJSON(out)
}

func cmdDiscover(args []string) error {
	fs, rootFlagVal := rootFlag(args, "discover")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	if err := ensureDirs(root); err != nil {
		return err
	}
	cfg, err := loadConfig(root)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	type fileInfo struct {
		name, hash string
		size       int64
		mod        string
	}
	files := []fileInfo{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasPrefix(n, "~$") {
			continue
		}
		if strings.ToLower(filepath.Ext(n)) != ".docx" {
			continue
		}
		full := filepath.Join(root, n)
		h, err := sha256File(full)
		if err != nil {
			return err
		}
		inf, err := e.Info()
		if err != nil {
			return err
		}
		files = append(files, fileInfo{name: n, hash: h, size: inf.Size(), mod: inf.ModTime().UTC().Format(time.RFC3339)})
	}
	sort.Slice(files, func(i, j int) bool { return strings.ToLower(files[i].name) < strings.ToLower(files[j].name) })

	byName := map[string]int{}
	byHash := map[string]int{}
	for i := range reg.Calls {
		byName[strings.ToLower(reg.Calls[i].SourceFile)] = i
		if reg.Calls[i].SourceSHA256 != "" {
			byHash[reg.Calls[i].SourceSHA256] = i
		}
		reg.Calls[i].SourceMissing = true
	}

	pending, unchanged, missing := []CallEntry{}, []CallEntry{}, []CallEntry{}
	duplicates := []DuplicateFinding{}
	for _, f := range files {
		key := strings.ToLower(f.name)
		if idx, ok := byName[key]; ok {
			c := &reg.Calls[idx]
			c.SourceMissing = false
			c.SourceSize = f.size
			c.SourceModifiedUTC = f.mod
			if c.SourceSHA256 == f.hash && c.Status == "complete" {
				unchanged = append(unchanged, *c)
			} else if c.SourceSHA256 == f.hash && c.Status != "failed" {
				if c.Status == "" {
					c.Status = "pending"
				}
				pending = append(pending, *c)
			} else {
				c.SourceSHA256 = f.hash
				c.Status = "pending"
				c.Error = ""
				pending = append(pending, *c)
				byHash[f.hash] = idx
			}
			continue
		}
		if idx, ok := byHash[f.hash]; ok {
			existing := &reg.Calls[idx]
			if existing.SourceMissing {
				delete(byName, strings.ToLower(existing.SourceFile))
				existing.SourceFile = f.name
				existing.SourceSize = f.size
				existing.SourceModifiedUTC = f.mod
				existing.SourceMissing = false
				byName[key] = idx
				if existing.Status == "complete" {
					unchanged = append(unchanged, *existing)
				} else {
					pending = append(pending, *existing)
				}
			} else {
				duplicates = append(duplicates, DuplicateFinding{SourceFile: f.name, DuplicateOf: existing.CallID, SHA256: f.hash})
			}
			continue
		}
		seq := reg.NextSequence
		reg.NextSequence++
		callID := fmt.Sprintf("%s-%04d", cfg.CallIDPrefix, seq)
		c := CallEntry{CallID: callID, Sequence: seq, SourceFile: f.name, SourceSHA256: f.hash, SourceSize: f.size, SourceModifiedUTC: f.mod, Status: "pending", FirstSeenUTC: nowUTC()}
		reg.Calls = append(reg.Calls, c)
		idx := len(reg.Calls) - 1
		byName[key] = idx
		byHash[f.hash] = idx
		pending = append(pending, c)
	}
	for i := range reg.Calls {
		if reg.Calls[i].SourceMissing {
			missing = append(missing, reg.Calls[i])
		}
	}
	sort.Slice(reg.Calls, func(i, j int) bool { return reg.Calls[i].Sequence < reg.Calls[j].Sequence })
	if err := saveRegistry(root, reg); err != nil {
		return err
	}
	out := DiscoverOutput{Pending: pending, Unchanged: unchanged, Missing: missing, Duplicates: duplicates, Summary: map[string]int{"pending": len(pending), "unchanged": len(unchanged), "missing": len(missing), "duplicates": len(duplicates), "total_registered": len(reg.Calls)}}
	return printJSON(out)
}

func findCall(reg *Registry, callID string) (*CallEntry, error) {
	for i := range reg.Calls {
		if reg.Calls[i].CallID == callID {
			return &reg.Calls[i], nil
		}
	}
	return nil, fmt.Errorf("call ID not found: %s", callID)
}

func cmdExtract(args []string) error {
	fs, rootFlagVal := rootFlag(args, "extract")
	callID := fs.String("call-id", "", "registered call ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *callID == "" {
		return errors.New("--call-id is required")
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	c, err := findCall(&reg, *callID)
	if err != nil {
		return err
	}
	source := filepath.Join(root, filepath.FromSlash(c.SourceFile))
	h, err := sha256File(source)
	if err != nil {
		return err
	}
	if h != c.SourceSHA256 {
		return fmt.Errorf("source changed after discovery; rerun discover for %s", c.SourceFile)
	}
	paras, err := extractDOCXParagraphs(source)
	if err != nil {
		return err
	}
	if len(paras) == 0 {
		return errors.New("no readable paragraphs found in DOCX")
	}
	outPath := filepath.Join(paths(root)["extracted"], c.CallID+".md")
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("call_id: " + c.CallID + "\n")
	b.WriteString("source_file: " + yamlQuote(c.SourceFile) + "\n")
	b.WriteString("source_sha256: " + c.SourceSHA256 + "\n")
	b.WriteString("paragraph_count: " + strconv.Itoa(len(paras)) + "\n")
	b.WriteString("---\n\n")
	for i, p := range paras {
		fmt.Fprintf(&b, "[P%04d] %s\n\n", i+1, strings.TrimSpace(p))
	}
	if err := writeTextAtomic(outPath, b.String()); err != nil {
		return err
	}
	c.ExtractedPath = relPath(root, outPath)
	c.SummarySpecPath = relPath(root, filepath.Join(paths(root)["specs"], c.CallID+".summary.json"))
	c.CallRecordPath = relPath(root, filepath.Join(paths(root)["records"], c.CallID+".record.json"))
	c.Status = "extracted"
	c.Error = ""
	if err := saveRegistry(root, reg); err != nil {
		return err
	}
	return printJSON(map[string]interface{}{
		"call_id":           c.CallID,
		"source_file":       c.SourceFile,
		"source_sha256":     c.SourceSHA256,
		"paragraph_count":   len(paras),
		"extracted_path":    c.ExtractedPath,
		"summary_spec_path": c.SummarySpecPath,
		"call_record_path":  c.CallRecordPath,
	})
}

func yamlQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func extractDOCXParagraphs(path string) ([]string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open DOCX: %w", err)
	}
	defer zr.Close()
	var doc *zip.File
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			doc = f
			break
		}
	}
	if doc == nil {
		return nil, errors.New("DOCX missing word/document.xml")
	}
	rc, err := doc.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	dec := xml.NewDecoder(rc)
	var paras []string
	var current strings.Builder
	inP := 0
	inDel := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				if inP == 0 {
					current.Reset()
				}
				inP++
			case "del":
				inDel++
			case "t":
				if inP > 0 && inDel == 0 {
					var txt string
					if err := dec.DecodeElement(&txt, &t); err != nil {
						return nil, err
					}
					current.WriteString(txt)
				}
			case "tab":
				if inP > 0 && inDel == 0 {
					current.WriteByte('\t')
				}
			case "br", "cr":
				if inP > 0 && inDel == 0 {
					current.WriteByte('\n')
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p":
				inP--
				if inP == 0 {
					s := cleanParagraphText(current.String())
					if s != "" {
						paras = append(paras, s)
					}
				}
			case "del":
				if inDel > 0 {
					inDel--
				}
			}
		}
	}
	return paras, nil
}

func cleanParagraphText(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.Join(strings.Fields(lines[i]), " ")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func cmdValidate(args []string) error {
	fs, rootFlagVal := rootFlag(args, "validate")
	callID := fs.String("call-id", "", "call ID")
	summaryPath := fs.String("summary", "", "summary spec path")
	recordPath := fs.String("record", "", "call record path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	if *callID == "" {
		return errors.New("--call-id is required")
	}
	c, err := findCall(&reg, *callID)
	if err != nil {
		return err
	}
	if *summaryPath == "" {
		*summaryPath = filepath.Join(root, filepath.FromSlash(c.SummarySpecPath))
	} else if !filepath.IsAbs(*summaryPath) {
		*summaryPath = filepath.Join(root, *summaryPath)
	}
	if *recordPath == "" {
		*recordPath = filepath.Join(root, filepath.FromSlash(c.CallRecordPath))
	} else if !filepath.IsAbs(*recordPath) {
		*recordPath = filepath.Join(root, *recordPath)
	}
	var spec SummarySpec
	if err := readJSON(*summaryPath, &spec); err != nil {
		return err
	}
	var rec CallRecord
	if err := readJSON(*recordPath, &rec); err != nil {
		return err
	}
	errs := append(validateSummarySpec(spec, *c), validateCallRecord(rec, *c, root)...)
	if len(errs) > 0 {
		return fmt.Errorf("validation failed:\n- %s", strings.Join(errs, "\n- "))
	}
	return printJSON(map[string]interface{}{"status": "valid", "call_id": *callID, "summary": relPath(root, *summaryPath), "record": relPath(root, *recordPath), "evidence_count": len(rec.Evidence), "theme_count": len(spec.Interviews[0].Themes)})
}

func validateSummarySpec(s SummarySpec, c CallEntry) []string {
	var errs []string
	if s.SchemaVersion == "" {
		errs = append(errs, "summary schema_version is required")
	}
	if s.CallID != c.CallID {
		errs = append(errs, "summary call_id must match registry")
	}
	if !strings.EqualFold(filepath.Clean(s.SourceFile), filepath.Clean(c.SourceFile)) {
		errs = append(errs, "summary source_file must match registry")
	}
	if s.SourceSHA256 != c.SourceSHA256 {
		errs = append(errs, "summary source_sha256 must match registry")
	}
	if len(s.Interviews) != 1 {
		errs = append(errs, "summary must contain exactly one interview")
	}
	if len(s.Interviews) == 1 {
		in := s.Interviews[0]
		if !strings.HasPrefix(strings.TrimSpace(in.Title), "Interview") {
			errs = append(errs, "interview title must start with 'Interview'")
		}
		haveStake, haveFocus := false, false
		for _, f := range in.Framing {
			l := strings.ToLower(strings.TrimSpace(f))
			if strings.HasPrefix(l, "stakeholder type:") {
				haveStake = true
			}
			if strings.HasPrefix(l, "focus:") {
				haveFocus = true
			}
		}
		if !haveStake {
			errs = append(errs, "framing must include 'Stakeholder type:'")
		}
		if !haveFocus {
			errs = append(errs, "framing must include 'Focus:'")
		}
		if len(in.Themes) == 0 {
			errs = append(errs, "at least one theme is required")
		}
		for i, t := range in.Themes {
			if strings.TrimSpace(t.Lead) == "" {
				errs = append(errs, fmt.Sprintf("theme %d lead is empty", i+1))
			}
			if len(t.Points) == 0 {
				errs = append(errs, fmt.Sprintf("theme %d has no points", i+1))
			}
			for _, p := range t.Points {
				lp := strings.ToLower(strings.TrimSpace(p))
				if strings.HasPrefix(lp, "q:") || strings.HasPrefix(lp, "a:") {
					errs = append(errs, fmt.Sprintf("theme %d contains Q&A formatting", i+1))
				}
			}
		}
	}
	return errs
}

var anchorRe = regexp.MustCompile(`^P(\d{4})(?:-P(\d{4}))?$`)

func validateCallRecord(r CallRecord, c CallEntry, root string) []string {
	var errs []string
	if r.SchemaVersion == "" {
		errs = append(errs, "record schema_version is required")
	}
	if r.CallID != c.CallID {
		errs = append(errs, "record call_id must match registry")
	}
	if !strings.EqualFold(filepath.Clean(r.SourceFile), filepath.Clean(c.SourceFile)) {
		errs = append(errs, "record source_file must match registry")
	}
	if r.SourceSHA256 != c.SourceSHA256 {
		errs = append(errs, "record source_sha256 must match registry")
	}
	if strings.TrimSpace(r.InterviewTitle) == "" {
		errs = append(errs, "record interview_title is required")
	}
	if strings.TrimSpace(r.StakeholderType) == "" {
		errs = append(errs, "record stakeholder_type is required")
	}
	if strings.TrimSpace(r.Focus) == "" {
		errs = append(errs, "record focus is required")
	}
	if len(r.Evidence) == 0 {
		errs = append(errs, "record must contain at least one evidence item")
	}
	seen := map[string]bool{}
	maxPara := maxParagraphNumber(filepath.Join(root, filepath.FromSlash(c.ExtractedPath)))
	allowedModules := map[string]bool{"supplier": true, "customer": true, "competitor": true, "cross-cutting": true}
	for i, e := range r.Evidence {
		if e.EvidenceID == "" {
			errs = append(errs, fmt.Sprintf("evidence %d missing evidence_id", i+1))
		} else {
			expectedPrefix := c.CallID + "-E"
			if !strings.HasPrefix(e.EvidenceID, expectedPrefix) {
				errs = append(errs, fmt.Sprintf("evidence %s must start with %s", e.EvidenceID, expectedPrefix))
			}
			if seen[e.EvidenceID] {
				errs = append(errs, fmt.Sprintf("duplicate evidence_id %s", e.EvidenceID))
			}
			seen[e.EvidenceID] = true
		}
		if strings.TrimSpace(e.Assertion) == "" {
			errs = append(errs, fmt.Sprintf("evidence %s has empty assertion", e.EvidenceID))
		}
		if len(e.Modules) == 0 {
			errs = append(errs, fmt.Sprintf("evidence %s needs at least one module", e.EvidenceID))
		}
		for _, m := range e.Modules {
			if !allowedModules[m] {
				errs = append(errs, fmt.Sprintf("evidence %s has invalid module %q", e.EvidenceID, m))
			}
		}
		if strings.TrimSpace(e.Topic) == "" {
			errs = append(errs, fmt.Sprintf("evidence %s needs topic", e.EvidenceID))
		}
		if len(e.SourceAnchors) == 0 {
			errs = append(errs, fmt.Sprintf("evidence %s needs source_anchors", e.EvidenceID))
		}
		for _, a := range e.SourceAnchors {
			m := anchorRe.FindStringSubmatch(strings.TrimSpace(a))
			if m == nil {
				errs = append(errs, fmt.Sprintf("evidence %s has invalid anchor %q", e.EvidenceID, a))
				continue
			}
			n1, _ := strconv.Atoi(m[1])
			n2 := n1
			if m[2] != "" {
				n2, _ = strconv.Atoi(m[2])
			}
			if n1 < 1 || n2 < n1 || (maxPara > 0 && n2 > maxPara) {
				errs = append(errs, fmt.Sprintf("evidence %s anchor %q is outside extracted transcript", e.EvidenceID, a))
			}
		}
	}
	return errs
}

func maxParagraphNumber(path string) int {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	re := regexp.MustCompile(`\[P(\d{4})\]`)
	max := 0
	for _, m := range re.FindAllStringSubmatch(string(b), -1) {
		n, _ := strconv.Atoi(m[1])
		if n > max {
			max = n
		}
	}
	return max
}

func cmdRenderSummary(args []string) error {
	fs, rootFlagVal := rootFlag(args, "render-summary")
	callID := fs.String("call-id", "", "call ID")
	specPath := fs.String("spec", "", "summary spec path")
	outPath := fs.String("out", "", "output DOCX path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	if *callID == "" {
		return errors.New("--call-id is required")
	}
	c, err := findCall(&reg, *callID)
	if err != nil {
		return err
	}
	if *specPath == "" {
		*specPath = filepath.Join(root, filepath.FromSlash(c.SummarySpecPath))
	} else if !filepath.IsAbs(*specPath) {
		*specPath = filepath.Join(root, *specPath)
	}
	var spec SummarySpec
	if err := readJSON(*specPath, &spec); err != nil {
		return err
	}
	if errs := validateSummarySpec(spec, *c); len(errs) > 0 {
		return fmt.Errorf("summary validation failed: %s", strings.Join(errs, "; "))
	}
	if *outPath == "" {
		*outPath = filepath.Join(paths(root)["processed"], fmt.Sprintf("%s_%s_Summary.docx", c.CallID, safeFilename(c.SourceFile)))
	} else if !filepath.IsAbs(*outPath) {
		*outPath = filepath.Join(root, *outPath)
	}
	if err := writeSummaryDOCX(*outPath, spec.Interviews, false); err != nil {
		return err
	}
	mdPath := strings.TrimSuffix(*outPath, filepath.Ext(*outPath)) + ".md"
	if err := writeTextAtomic(mdPath, summaryMarkdown(spec.Interviews)); err != nil {
		return err
	}
	c.SummaryDocxPath = relPath(root, *outPath)
	c.SummaryMarkdownPath = relPath(root, mdPath)
	if c.Status != "complete" {
		c.Status = "rendered"
	}
	if err := saveRegistry(root, reg); err != nil {
		return err
	}
	return printJSON(map[string]interface{}{"status": "rendered", "call_id": c.CallID, "summary_docx_path": c.SummaryDocxPath, "summary_markdown_path": c.SummaryMarkdownPath})
}

func summaryMarkdown(interviews []Interview) string {
	var b strings.Builder
	for i, in := range interviews {
		if i > 0 {
			b.WriteString("\n---\n\n")
		}
		b.WriteString("## " + in.Title + "\n\n")
		for _, f := range in.Framing {
			b.WriteString("- " + f + "\n")
		}
		b.WriteString("\n")
		for _, t := range in.Themes {
			b.WriteString("- **" + t.Lead + "**\n")
			for _, p := range t.Points {
				b.WriteString("  - " + p + "\n")
			}
		}
	}
	return b.String()
}

func cmdMarkComplete(args []string) error {
	fs, rootFlagVal := rootFlag(args, "mark-complete")
	callID := fs.String("call-id", "", "call ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	if *callID == "" {
		return errors.New("--call-id is required")
	}
	c, err := findCall(&reg, *callID)
	if err != nil {
		return err
	}
	required := []string{c.ExtractedPath, c.SummarySpecPath, c.CallRecordPath, c.SummaryDocxPath, c.SummaryMarkdownPath}
	for _, rp := range required {
		if rp == "" {
			return fmt.Errorf("call %s is missing an output path", c.CallID)
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rp))); err != nil {
			return fmt.Errorf("required output missing %s: %w", rp, err)
		}
	}
	c.Status = "complete"
	c.LastProcessedUTC = nowUTC()
	c.Error = ""
	c.SourceMissing = false
	if err := saveRegistry(root, reg); err != nil {
		return err
	}
	return printJSON(map[string]interface{}{"status": "complete", "call_id": c.CallID, "sequence": c.Sequence})
}

func cmdMarkFailed(args []string) error {
	fs, rootFlagVal := rootFlag(args, "mark-failed")
	callID := fs.String("call-id", "", "call ID")
	message := fs.String("message", "unspecified processing failure", "failure message")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	c, err := findCall(&reg, *callID)
	if err != nil {
		return err
	}
	c.Status = "failed"
	c.Error = *message
	if err := saveRegistry(root, reg); err != nil {
		return err
	}
	reviewPath := filepath.Join(paths(root)["review"], c.CallID+"_FAILED.md")
	text := fmt.Sprintf("# %s processing failure\n\n- Source: `%s`\n- Time: %s\n- Error: %s\n", c.CallID, c.SourceFile, nowUTC(), *message)
	_ = writeTextAtomic(reviewPath, text)
	return printJSON(map[string]interface{}{"status": "failed", "call_id": c.CallID, "message": *message})
}

func cmdBuildLedger(args []string) error {
	fs, rootFlagVal := rootFlag(args, "build-ledger")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	if err := ensureDirs(root); err != nil {
		return err
	}
	cfg, err := loadConfig(root)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	var rows []EvidenceLedgerRow
	var sources []SourceRef
	entityMap := map[string]*EntityAggregate{}
	openQuestions := []struct {
		CallID     string
		SourceFile string
		Q          OpenQuestion
	}{}
	for _, c := range reg.Calls {
		if c.Status != "complete" {
			continue
		}
		var rec CallRecord
		if err := readJSON(filepath.Join(root, filepath.FromSlash(c.CallRecordPath)), &rec); err != nil {
			return err
		}
		if errs := validateCallRecord(rec, c, root); len(errs) > 0 {
			return fmt.Errorf("record %s invalid: %s", c.CallID, strings.Join(errs, "; "))
		}
		src := SourceRef{CallID: c.CallID, Sequence: c.Sequence, SourceFile: c.SourceFile, SourceSHA256: c.SourceSHA256, InterviewTitle: rec.InterviewTitle, StakeholderType: rec.StakeholderType, Focus: rec.Focus, ExpertName: rec.Interviewee.Name, ExpertRole: rec.Interviewee.Role, ExpertCompany: rec.Interviewee.Company, ExtractedPath: c.ExtractedPath, SummaryDocxPath: c.SummaryDocxPath, SummaryMarkdownPath: c.SummaryMarkdownPath}
		sources = append(sources, src)
		for _, e := range rec.Evidence {
			rows = append(rows, EvidenceLedgerRow{Evidence: e, Source: src})
		}
		for _, en := range rec.Entities {
			key := slugify(firstNonEmpty(en.CanonicalName, en.DisplayName))
			if _, ok := entityMap[key]; !ok {
				entityMap[key] = &EntityAggregate{CanonicalName: key, DisplayName: firstNonEmpty(en.DisplayName, en.CanonicalName), EntityType: en.EntityType, Modules: map[string]int{}, Topics: map[string]int{}}
			}
			a := entityMap[key]
			a.Roles = uniqueSorted(append(a.Roles, en.Roles...))
			a.Aliases = uniqueSorted(append(a.Aliases, en.Aliases...))
			a.SourceCalls = uniqueSorted(append(a.SourceCalls, c.CallID))
		}
		for _, q := range rec.OpenQuestions {
			openQuestions = append(openQuestions, struct {
				CallID     string
				SourceFile string
				Q          OpenQuestion
			}{c.CallID, c.SourceFile, q})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Source.Sequence != rows[j].Source.Sequence {
			return rows[i].Source.Sequence < rows[j].Source.Sequence
		}
		return rows[i].Evidence.EvidenceID < rows[j].Evidence.EvidenceID
	})
	sort.Slice(sources, func(i, j int) bool { return sources[i].Sequence < sources[j].Sequence })
	kdir := paths(root)["knowledge"]
	if err := writeJSONL(filepath.Join(kdir, "evidence_ledger.jsonl"), rows); err != nil {
		return err
	}
	if err := writeJSONAtomic(filepath.Join(kdir, "source_registry.json"), sources); err != nil {
		return err
	}
	if err := writeOpenQuestions(filepath.Join(kdir, "OPEN_QUESTIONS.md"), openQuestions); err != nil {
		return err
	}

	groups := map[string][]EvidenceLedgerRow{}
	groupDefs := map[string]JobGroup{}
	for _, row := range rows {
		module := "cross-cutting"
		if len(row.Evidence.Modules) > 0 {
			module = row.Evidence.Modules[0]
		}
		topic := slugify(row.Evidence.Topic)
		entity := slugify(row.Evidence.Subject)
		key := module + "|" + topic + "|" + entity
		groups[key] = append(groups[key], row)
		groupDefs[key] = JobGroup{Module: module, Topic: topic, PrimaryEntity: entity}
		if entity != "general" {
			if _, ok := entityMap[entity]; !ok {
				entityMap[entity] = &EntityAggregate{CanonicalName: entity, DisplayName: row.Evidence.Subject, Modules: map[string]int{}, Topics: map[string]int{}}
			}
			a := entityMap[entity]
			a.EvidenceCount++
			a.Modules[module]++
			a.Topics[topic]++
			a.SourceCalls = uniqueSorted(append(a.SourceCalls, row.Source.CallID))
		}
	}
	var keys []string
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	manifest := CompileJobManifest{SchemaVersion: "1.0", GeneratedUTC: nowUTC()}
	for _, k := range keys {
		evs := groups[k]
		if len(evs) > cfg.MaxCompileJobItems {
			return fmt.Errorf("compile group %s has %d items, exceeding configured max %d; narrow the ontology or raise max_compile_job_items", k, len(evs), cfg.MaxCompileJobItems)
		}
		g := groupDefs[k]
		raw, _ := json.Marshal(struct {
			Group    JobGroup            `json:"group"`
			Evidence []EvidenceLedgerRow `json:"evidence"`
		}{g, evs})
		inputHash := hashBytes(raw)
		jobID := "JOB-" + strings.ToUpper(inputHash[:12])
		jobRel := filepath.ToSlash(filepath.Join(outputRootName, "_System", "compile_jobs", jobID+".json"))
		outRel := filepath.ToSlash(filepath.Join(outputRootName, "_System", "claim_shards", jobID+".json"))
		job := CompileJob{SchemaVersion: "1.0", JobID: jobID, InputHash: inputHash, Group: g, OutputPath: outRel, Evidence: evs}
		if err := writeJSONAtomic(filepath.Join(root, filepath.FromSlash(jobRel)), job); err != nil {
			return err
		}
		status := "pending"
		var shard ClaimShard
		if err := readJSON(filepath.Join(root, filepath.FromSlash(outRel)), &shard); err == nil && shard.JobID == jobID && shard.InputHash == inputHash {
			status = "complete"
		}
		manifest.Jobs = append(manifest.Jobs, CompileJobManifestEntry{JobID: jobID, InputHash: inputHash, Group: g, JobPath: jobRel, OutputPath: outRel, Status: status, ItemCount: len(evs)})
	}
	if err := writeJSONAtomic(paths(root)["manifest"], manifest); err != nil {
		return err
	}
	entities := make([]EntityAggregate, 0, len(entityMap))
	for _, v := range entityMap {
		entities = append(entities, *v)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].CanonicalName < entities[j].CanonicalName })
	if err := writeJSONAtomic(filepath.Join(kdir, "entity_registry.json"), entities); err != nil {
		return err
	}
	return printJSON(map[string]interface{}{"status": "ledger_built", "calls": len(sources), "evidence": len(rows), "entities": len(entities), "compile_jobs": len(manifest.Jobs), "pending_jobs": countPendingJobs(manifest)})
}

func firstNonEmpty(xs ...string) string {
	for _, x := range xs {
		if strings.TrimSpace(x) != "" {
			return strings.TrimSpace(x)
		}
	}
	return "general"
}

func writeJSONL[T any](path string, rows []T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return writeTextAtomic(path, b.String())
}

func readJSONL[T any](path string) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []T
	s := bufio.NewScanner(f)
	buf := make([]byte, 1024*1024)
	s.Buffer(buf, 16*1024*1024)
	line := 0
	for s.Scan() {
		line++
		t := strings.TrimSpace(s.Text())
		if t == "" {
			continue
		}
		var v T
		if err := json.Unmarshal([]byte(t), &v); err != nil {
			return nil, fmt.Errorf("%s line %d: %w", path, line, err)
		}
		out = append(out, v)
	}
	return out, s.Err()
}

func writeOpenQuestions(path string, qs []struct {
	CallID     string
	SourceFile string
	Q          OpenQuestion
}) error {
	var b strings.Builder
	b.WriteString("# Open Questions\n\n")
	if len(qs) == 0 {
		b.WriteString("No open questions have been extracted yet\n")
		return writeTextAtomic(path, b.String())
	}
	sort.Slice(qs, func(i, j int) bool {
		if qs[i].Q.Priority != qs[j].Q.Priority {
			return qs[i].Q.Priority < qs[j].Q.Priority
		}
		return qs[i].CallID < qs[j].CallID
	})
	for _, x := range qs {
		fmt.Fprintf(&b, "- **%s**: %s\n  - Source: `%s`\n", x.CallID, x.Q.Question, x.SourceFile)
		if x.Q.WhyItMatters != "" {
			fmt.Fprintf(&b, "  - Why it matters: %s\n", x.Q.WhyItMatters)
		}
		if x.Q.Priority != "" {
			fmt.Fprintf(&b, "  - Priority: %s\n", x.Q.Priority)
		}
	}
	return writeTextAtomic(path, b.String())
}

func countPendingJobs(m CompileJobManifest) int {
	n := 0
	for _, j := range m.Jobs {
		if j.Status != "complete" {
			n++
		}
	}
	return n
}

func cmdDiscoverJobs(args []string) error {
	fs, rootFlagVal := rootFlag(args, "discover-jobs")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	var m CompileJobManifest
	if err := readJSON(paths(root)["manifest"], &m); err != nil {
		return fmt.Errorf("run build-ledger first: %w", err)
	}
	pending := []CompileJobManifestEntry{}
	complete := 0
	for _, j := range m.Jobs {
		if j.Status == "complete" {
			complete++
		} else {
			pending = append(pending, j)
		}
	}
	return printJSON(map[string]interface{}{"pending": pending, "summary": map[string]int{"pending": len(pending), "complete": complete, "total": len(m.Jobs)}})
}

func cmdValidateShard(args []string) error {
	fs, rootFlagVal := rootFlag(args, "validate-shard")
	jobPath := fs.String("job", "", "compile job path")
	shardPath := fs.String("shard", "", "claim shard path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	if *jobPath == "" || *shardPath == "" {
		return errors.New("--job and --shard are required")
	}
	if !filepath.IsAbs(*jobPath) {
		*jobPath = filepath.Join(root, filepath.FromSlash(*jobPath))
	}
	if !filepath.IsAbs(*shardPath) {
		*shardPath = filepath.Join(root, filepath.FromSlash(*shardPath))
	}
	var job CompileJob
	if err := readJSON(*jobPath, &job); err != nil {
		return err
	}
	var shard ClaimShard
	if err := readJSON(*shardPath, &shard); err != nil {
		return err
	}
	errs := validateShard(job, shard)
	if len(errs) > 0 {
		return fmt.Errorf("claim shard validation failed:\n- %s", strings.Join(errs, "\n- "))
	}
	var manifest CompileJobManifest
	if err := readJSON(paths(root)["manifest"], &manifest); err == nil {
		changed := false
		for i := range manifest.Jobs {
			if manifest.Jobs[i].JobID == job.JobID {
				manifest.Jobs[i].Status = "complete"
				changed = true
				break
			}
		}
		if changed {
			if err := writeJSONAtomic(paths(root)["manifest"], manifest); err != nil {
				return err
			}
		}
	}
	return printJSON(map[string]interface{}{"status": "valid", "job_id": job.JobID, "claims": len(shard.Claims), "unclustered": len(shard.UnclusteredEvidence)})
}

func validateShard(job CompileJob, shard ClaimShard) []string {
	var errs []string
	if shard.SchemaVersion == "" {
		errs = append(errs, "schema_version is required")
	}
	if shard.JobID != job.JobID {
		errs = append(errs, "job_id does not match job")
	}
	if shard.InputHash != job.InputHash {
		errs = append(errs, "input_hash does not match job")
	}
	allowed := map[string]bool{}
	for _, r := range job.Evidence {
		allowed[r.Evidence.EvidenceID] = true
	}
	covered := map[string]bool{}
	for i, c := range shard.Claims {
		if strings.TrimSpace(c.ClaimKey) == "" {
			errs = append(errs, fmt.Sprintf("claim %d missing claim_key", i+1))
		}
		if strings.TrimSpace(c.NormalizedClaim) == "" {
			errs = append(errs, fmt.Sprintf("claim %d missing normalized_claim", i+1))
		}
		if strings.TrimSpace(c.Synthesis) == "" {
			errs = append(errs, fmt.Sprintf("claim %d missing synthesis", i+1))
		}
		all := [][]string{c.SupportingEvidence, c.ContradictingEvidence, c.QualifyingEvidence}
		for _, lst := range all {
			for _, id := range lst {
				if !allowed[id] {
					errs = append(errs, fmt.Sprintf("claim %s references unknown evidence %s", c.ClaimKey, id))
				} else {
					covered[id] = true
				}
			}
		}
	}
	for _, id := range shard.UnclusteredEvidence {
		if !allowed[id] {
			errs = append(errs, fmt.Sprintf("unknown unclustered evidence %s", id))
		} else {
			covered[id] = true
		}
	}
	for id := range allowed {
		if !covered[id] {
			errs = append(errs, fmt.Sprintf("evidence %s is neither clustered nor unclustered", id))
		}
	}
	return errs
}

func cmdFinalizeIndex(args []string) error {
	fs, rootFlagVal := rootFlag(args, "finalize-index")
	allowIncomplete := fs.Bool("allow-incomplete", false, "build maps with missing shards")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	var manifest CompileJobManifest
	if err := readJSON(paths(root)["manifest"], &manifest); err != nil {
		return fmt.Errorf("run build-ledger first: %w", err)
	}
	var claims []FinalClaim
	missing := []string{}
	for _, m := range manifest.Jobs {
		var job CompileJob
		if err := readJSON(filepath.Join(root, filepath.FromSlash(m.JobPath)), &job); err != nil {
			return err
		}
		var shard ClaimShard
		sp := filepath.Join(root, filepath.FromSlash(m.OutputPath))
		if err := readJSON(sp, &shard); err != nil {
			missing = append(missing, m.JobID)
			continue
		}
		if errs := validateShard(job, shard); len(errs) > 0 {
			return fmt.Errorf("invalid shard %s: %s", m.JobID, strings.Join(errs, "; "))
		}
		for _, c := range shard.Claims {
			key := c.ClaimKey
			if key == "" {
				key = slugify(c.NormalizedClaim)
			}
			scopeBytes, _ := json.Marshal(c.Scope)
			idHash := hashString(strings.Join([]string{job.Group.Module, job.Group.Topic, job.Group.PrimaryEntity, key, string(scopeBytes)}, "|"))
			claims = append(claims, FinalClaim{ClaimID: "CLM-" + strings.ToUpper(idHash[:12]), ClaimKey: key, NormalizedClaim: c.NormalizedClaim, Synthesis: c.Synthesis, Status: c.Status, Confidence: c.Confidence, Module: job.Group.Module, Topic: job.Group.Topic, PrimaryEntity: job.Group.PrimaryEntity, PrimaryEntities: c.PrimaryEntities, Scope: c.Scope, SupportingEvidence: uniqueSorted(c.SupportingEvidence), ContradictingEvidence: uniqueSorted(c.ContradictingEvidence), QualifyingEvidence: uniqueSorted(c.QualifyingEvidence), OpenQuestions: c.OpenQuestions})
		}
	}
	if len(missing) > 0 && !*allowIncomplete {
		return fmt.Errorf("missing claim shards for %d jobs: %s", len(missing), strings.Join(missing, ", "))
	}
	sort.Slice(claims, func(i, j int) bool {
		if claims[i].Module != claims[j].Module {
			return claims[i].Module < claims[j].Module
		}
		if claims[i].Topic != claims[j].Topic {
			return claims[i].Topic < claims[j].Topic
		}
		if claims[i].PrimaryEntity != claims[j].PrimaryEntity {
			return claims[i].PrimaryEntity < claims[j].PrimaryEntity
		}
		return claims[i].ClaimID < claims[j].ClaimID
	})
	kdir := paths(root)["knowledge"]
	if err := writeJSONL(filepath.Join(kdir, "claim_registry.jsonl"), claims); err != nil {
		return err
	}
	rows, err := readJSONL[EvidenceLedgerRow](filepath.Join(kdir, "evidence_ledger.jsonl"))
	if err != nil {
		return err
	}
	if err := buildMapsAndCards(root, claims, rows); err != nil {
		return err
	}
	return printJSON(map[string]interface{}{"status": "index_finalized", "claims": len(claims), "missing_shards": missing, "case_map": relPath(root, filepath.Join(kdir, "CASE_MAP.md"))})
}

func buildMapsAndCards(root string, claims []FinalClaim, rows []EvidenceLedgerRow) error {
	kdir := paths(root)["knowledge"]
	entityMap := map[string]*EntityAggregate{}
	var existing []EntityAggregate
	_ = readJSON(filepath.Join(kdir, "entity_registry.json"), &existing)
	for i := range existing {
		e := existing[i]
		if e.Modules == nil {
			e.Modules = map[string]int{}
		}
		if e.Topics == nil {
			e.Topics = map[string]int{}
		}
		entityMap[e.CanonicalName] = &e
	}
	topicCounts := map[string]map[string]int{}
	moduleCounts := map[string]int{}
	statusCounts := map[string]int{}
	for _, c := range claims {
		moduleCounts[c.Module]++
		statusCounts[strings.ToLower(c.Status)]++
		if topicCounts[c.Module] == nil {
			topicCounts[c.Module] = map[string]int{}
		}
		topicCounts[c.Module][c.Topic]++
		ent := c.PrimaryEntity
		if ent != "" && ent != "general" {
			if entityMap[ent] == nil {
				entityMap[ent] = &EntityAggregate{CanonicalName: ent, DisplayName: ent, Modules: map[string]int{}, Topics: map[string]int{}}
			}
			entityMap[ent].ClaimCount++
			entityMap[ent].Modules[c.Module]++
			entityMap[ent].Topics[c.Topic]++
		}
	}
	entities := make([]EntityAggregate, 0, len(entityMap))
	for _, e := range entityMap {
		entities = append(entities, *e)
	}
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].ClaimCount != entities[j].ClaimCount {
			return entities[i].ClaimCount > entities[j].ClaimCount
		}
		return entities[i].CanonicalName < entities[j].CanonicalName
	})
	if err := writeJSONAtomic(filepath.Join(kdir, "entity_registry.json"), entities); err != nil {
		return err
	}
	topicIndex := map[string]interface{}{"generated_utc": nowUTC(), "module_counts": moduleCounts, "topic_counts": topicCounts, "status_counts": statusCounts}
	if err := writeJSONAtomic(filepath.Join(kdir, "topic_index.json"), topicIndex); err != nil {
		return err
	}

	sourcesMap := map[string]SourceRef{}
	for _, r := range rows {
		sourcesMap[r.Source.CallID] = r.Source
	}
	var sources []SourceRef
	for _, s := range sourcesMap {
		sources = append(sources, s)
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Sequence < sources[j].Sequence })
	var caseMap strings.Builder
	caseMap.WriteString("# Case Map\n\n")
	fmt.Fprintf(&caseMap, "Updated: %s\n\n", nowUTC())
	fmt.Fprintf(&caseMap, "- Calls processed: %d\n- Atomic evidence records: %d\n- Normalized claim clusters: %d\n", len(sources), len(rows), len(claims))
	caseMap.WriteString("\n## How to use this map\n\nStart here, then use the case-search utility to retrieve only the relevant claims and evidence. Do not load every transcript into context.\n\n")
	caseMap.WriteString("## Modules\n\n")
	modules := []string{"supplier", "customer", "competitor", "cross-cutting"}
	for _, m := range modules {
		fmt.Fprintf(&caseMap, "- **%s**: %d claims across %d topics\n", titleCase(m), moduleCounts[m], len(topicCounts[m]))
	}
	caseMap.WriteString("\n## Most contested or mixed claims\n\n")
	contested := filterClaims(claims, func(c FinalClaim) bool {
		s := strings.ToLower(c.Status)
		return strings.Contains(s, "contested") || strings.Contains(s, "mixed") || len(c.ContradictingEvidence) > 0
	})
	sort.Slice(contested, func(i, j int) bool {
		return len(contested[i].ContradictingEvidence) > len(contested[j].ContradictingEvidence)
	})
	if len(contested) == 0 {
		caseMap.WriteString("- None yet\n")
	} else {
		for i, c := range contested {
			if i >= 20 {
				break
			}
			fmt.Fprintf(&caseMap, "- `%s` **%s**: %s (%d supporting, %d contradicting, %d qualifying)\n", c.ClaimID, c.Status, c.NormalizedClaim, len(c.SupportingEvidence), len(c.ContradictingEvidence), len(c.QualifyingEvidence))
		}
	}
	caseMap.WriteString("\n## Evidence gaps\n\n")
	gaps := filterClaims(claims, func(c FinalClaim) bool {
		return strings.Contains(strings.ToLower(c.Status), "single") || strings.EqualFold(c.Confidence, "low") || len(c.OpenQuestions) > 0
	})
	if len(gaps) == 0 {
		caseMap.WriteString("- None explicitly flagged\n")
	} else {
		for i, c := range gaps {
			if i >= 20 {
				break
			}
			fmt.Fprintf(&caseMap, "- `%s` %s\n", c.ClaimID, c.NormalizedClaim)
		}
	}
	caseMap.WriteString("\n## High-coverage entities\n\n")
	for i, e := range entities {
		if i >= 25 {
			break
		}
		fmt.Fprintf(&caseMap, "- **%s**: %d claims, %d evidence records, %d calls\n", e.DisplayName, e.ClaimCount, e.EvidenceCount, len(e.SourceCalls))
	}
	caseMap.WriteString("\n## Topic routes\n\n")
	for _, m := range modules {
		caseMap.WriteString("### " + titleCase(m) + "\n\n")
		type kv struct {
			k string
			v int
		}
		var arr []kv
		for k, v := range topicCounts[m] {
			arr = append(arr, kv{k, v})
		}
		sort.Slice(arr, func(i, j int) bool {
			if arr[i].v != arr[j].v {
				return arr[i].v > arr[j].v
			}
			return arr[i].k < arr[j].k
		})
		if len(arr) == 0 {
			caseMap.WriteString("- No indexed topics yet\n")
		} else {
			for _, x := range arr {
				fmt.Fprintf(&caseMap, "- `%s`: %d claims\n", x.k, x.v)
			}
		}
		caseMap.WriteString("\n")
	}
	if err := writeTextAtomic(filepath.Join(kdir, "CASE_MAP.md"), caseMap.String()); err != nil {
		return err
	}

	for _, m := range modules {
		if err := writeModuleMap(filepath.Join(kdir, strings.ToUpper(strings.ReplaceAll(m, "-", "_"))+"_MAP.md"), m, claims); err != nil {
			return err
		}
	}
	if err := writeEntityCards(filepath.Join(kdir, "Entities"), entities, claims); err != nil {
		return err
	}
	if err := writeSourceCards(filepath.Join(kdir, "Sources"), sources, rows); err != nil {
		return err
	}
	return nil
}

func filterClaims(cs []FinalClaim, p func(FinalClaim) bool) []FinalClaim {
	var out []FinalClaim
	for _, c := range cs {
		if p(c) {
			out = append(out, c)
		}
	}
	return out
}
func titleCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '-' || r == '_' })
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, " ")
}

func writeModuleMap(path, module string, claims []FinalClaim) error {
	var b strings.Builder
	b.WriteString("# " + titleCase(module) + " Map\n\n")
	byTopic := map[string][]FinalClaim{}
	for _, c := range claims {
		if c.Module == module {
			byTopic[c.Topic] = append(byTopic[c.Topic], c)
		}
	}
	var topics []string
	for t := range byTopic {
		topics = append(topics, t)
	}
	sort.Strings(topics)
	if len(topics) == 0 {
		b.WriteString("No claims indexed yet\n")
		return writeTextAtomic(path, b.String())
	}
	for _, t := range topics {
		fmt.Fprintf(&b, "## %s\n\n", titleCase(t))
		cs := byTopic[t]
		sort.Slice(cs, func(i, j int) bool { return cs[i].ClaimID < cs[j].ClaimID })
		for _, c := range cs {
			fmt.Fprintf(&b, "- `%s` **%s**: %s\n", c.ClaimID, c.Status, c.Synthesis)
			if len(c.ContradictingEvidence) > 0 {
				fmt.Fprintf(&b, "  - Competing evidence: %d items\n", len(c.ContradictingEvidence))
			}
		}
		b.WriteString("\n")
	}
	return writeTextAtomic(path, b.String())
}

func writeEntityCards(dir string, entities []EntityAggregate, claims []FinalClaim) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for _, e := range entities {
		if e.CanonicalName == "general" {
			continue
		}
		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n", e.DisplayName)
		if len(e.Roles) > 0 {
			fmt.Fprintf(&b, "- Roles: %s\n", strings.Join(e.Roles, ", "))
		}
		fmt.Fprintf(&b, "- Source calls: %s\n- Evidence records: %d\n- Claim clusters: %d\n\n", strings.Join(e.SourceCalls, ", "), e.EvidenceCount, e.ClaimCount)
		b.WriteString("## Claims\n\n")
		n := 0
		for _, c := range claims {
			if c.PrimaryEntity == e.CanonicalName || containsFold(c.PrimaryEntities, e.DisplayName) || containsFold(c.PrimaryEntities, e.CanonicalName) {
				fmt.Fprintf(&b, "- `%s` **%s**: %s\n", c.ClaimID, c.Status, c.Synthesis)
				n++
			}
		}
		if n == 0 {
			b.WriteString("- No normalized claims yet\n")
		}
		if err := writeTextAtomic(filepath.Join(dir, e.CanonicalName+".md"), b.String()); err != nil {
			return err
		}
	}
	return nil
}

func writeSourceCards(dir string, sources []SourceRef, rows []EvidenceLedgerRow) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	byCall := map[string][]Evidence{}
	for _, r := range rows {
		byCall[r.Source.CallID] = append(byCall[r.Source.CallID], r.Evidence)
	}
	for _, s := range sources {
		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n", s.CallID)
		fmt.Fprintf(&b, "- Source: `%s`\n- Interview: %s\n- Stakeholder type: %s\n- Focus: %s\n", s.SourceFile, s.InterviewTitle, s.StakeholderType, s.Focus)
		if s.ExpertRole != "" || s.ExpertCompany != "" {
			fmt.Fprintf(&b, "- Expert scope: %s, %s\n", s.ExpertRole, s.ExpertCompany)
		}
		b.WriteString("\n## Atomic evidence\n\n")
		for _, e := range byCall[s.CallID] {
			fmt.Fprintf(&b, "- `%s` %s\n  - Topic: `%s`; modules: %s; confidence: %s\n  - Source anchors: %s\n", e.EvidenceID, e.Assertion, e.Topic, strings.Join(e.Modules, ", "), e.Confidence, strings.Join(e.SourceAnchors, ", "))
		}
		if err := writeTextAtomic(filepath.Join(dir, s.CallID+".md"), b.String()); err != nil {
			return err
		}
	}
	return nil
}

func containsFold(xs []string, s string) bool {
	for _, x := range xs {
		if strings.EqualFold(x, s) {
			return true
		}
	}
	return false
}

func cmdRenderMaster(args []string) error {
	fs, rootFlagVal := rootFlag(args, "render-master")
	outPath := fs.String("out", "", "master DOCX path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	cfg, err := loadConfig(root)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	type numbered struct {
		Seq int
		In  Interview
	}
	var all []numbered
	for _, c := range reg.Calls {
		if c.Status != "complete" {
			continue
		}
		var spec SummarySpec
		if err := readJSON(filepath.Join(root, filepath.FromSlash(c.SummarySpecPath)), &spec); err != nil {
			return err
		}
		for _, in := range spec.Interviews {
			in.Title = numberInterviewTitle(in.Title, c.Sequence)
			all = append(all, numbered{c.Sequence, in})
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Seq < all[j].Seq })
	interviews := make([]Interview, len(all))
	for i, x := range all {
		interviews[i] = x.In
	}
	if *outPath == "" {
		*outPath = filepath.Join(paths(root)["compiled"], cfg.MasterFilename)
	} else if !filepath.IsAbs(*outPath) {
		*outPath = filepath.Join(root, *outPath)
	}
	if err := writeSummaryDOCX(*outPath, interviews, true); err != nil {
		return err
	}
	mdPath := strings.TrimSuffix(*outPath, filepath.Ext(*outPath)) + ".md"
	if err := writeTextAtomic(mdPath, summaryMarkdown(interviews)); err != nil {
		return err
	}
	return printJSON(map[string]interface{}{"status": "master_rendered", "interviews": len(interviews), "master_docx_path": relPath(root, *outPath), "master_markdown_path": relPath(root, mdPath)})
}

var interviewPrefixRe = regexp.MustCompile(`^Interview(?:\s+\d+)?\s*[—-]\s*`)

func numberInterviewTitle(title string, seq int) string {
	rest := interviewPrefixRe.ReplaceAllString(strings.TrimSpace(title), "")
	if rest == "" {
		rest = strings.TrimSpace(title)
	}
	return fmt.Sprintf("Interview %d — %s", seq, rest)
}

func cmdSearch(args []string) error {
	fs, rootFlagVal := rootFlag(args, "search")
	query := fs.String("query", "", "question to retrieve")
	queryFile := fs.String("query-file", "", "UTF-8 file containing the question")
	outPath := fs.String("out", "", "optional Markdown output path")
	topClaims := fs.Int("top-claims", 12, "maximum claim clusters")
	topEvidence := fs.Int("top-evidence", 30, "maximum evidence rows")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*query) == "" && strings.TrimSpace(*queryFile) != "" {
		qf := *queryFile
		if !filepath.IsAbs(qf) {
			qf = filepath.Join(root, filepath.FromSlash(qf))
		}
		b, readErr := os.ReadFile(qf)
		if readErr != nil {
			return fmt.Errorf("read query file: %w", readErr)
		}
		*query = strings.TrimSpace(string(b))
	}
	if strings.TrimSpace(*query) == "" {
		return errors.New("--query or --query-file is required")
	}
	kdir := paths(root)["knowledge"]
	claims, err := readJSONL[FinalClaim](filepath.Join(kdir, "claim_registry.jsonl"))
	if err != nil {
		return fmt.Errorf("case index not ready; run compile-case: %w", err)
	}
	rows, err := readJSONL[EvidenceLedgerRow](filepath.Join(kdir, "evidence_ledger.jsonl"))
	if err != nil {
		return err
	}
	packet := buildQueryPacket(*query, claims, rows, *topClaims, *topEvidence)
	if *outPath != "" {
		if !filepath.IsAbs(*outPath) {
			*outPath = filepath.Join(root, *outPath)
		}
		if err := writeTextAtomic(*outPath, packet); err != nil {
			return err
		}
		return printJSON(map[string]interface{}{"status": "query_packet_written", "path": relPath(root, *outPath)})
	}
	fmt.Print(packet)
	return nil
}

type scoredClaim struct {
	C     FinalClaim
	Score float64
}
type scoredEvidence struct {
	R     EvidenceLedgerRow
	Score float64
}

func buildQueryPacket(q string, claims []FinalClaim, rows []EvidenceLedgerRow, topC, topE int) string {
	qTokens := expandQueryTokens(tokenize(q))
	docFreq := map[string]int{}
	claimTokens := make([]map[string]int, len(claims))
	for i, c := range claims {
		t := tokenCounts(strings.Join([]string{c.NormalizedClaim, c.Synthesis, c.Module, c.Topic, c.PrimaryEntity, strings.Join(c.PrimaryEntities, " ")}, " "))
		claimTokens[i] = t
		for k := range t {
			docFreq[k]++
		}
	}
	var sc []scoredClaim
	lowerQ := strings.ToLower(q)
	wantConflict := strings.Contains(lowerQ, "competing") || strings.Contains(lowerQ, "disagree") || strings.Contains(lowerQ, "contradict") || strings.Contains(lowerQ, "different views")
	for i, c := range claims {
		score := bm25Like(qTokens, claimTokens[i], docFreq, len(claims))
		text := strings.ToLower(c.NormalizedClaim + " " + c.Synthesis)
		if strings.Contains(text, strings.ToLower(strings.TrimSpace(q))) {
			score += 12
		}
		for _, t := range qTokens {
			if t == c.Module || t == c.Topic || t == c.PrimaryEntity {
				score += 3
			}
		}
		if wantConflict && (len(c.ContradictingEvidence) > 0 || strings.Contains(strings.ToLower(c.Status), "mixed") || strings.Contains(strings.ToLower(c.Status), "contested")) {
			score += 8
		}
		if score > 0 {
			sc = append(sc, scoredClaim{c, score})
		}
	}
	sort.Slice(sc, func(i, j int) bool {
		if sc[i].Score != sc[j].Score {
			return sc[i].Score > sc[j].Score
		}
		return sc[i].C.ClaimID < sc[j].C.ClaimID
	})
	if len(sc) > topC {
		sc = sc[:topC]
	}
	selected := map[string]float64{}
	for rank, x := range sc {
		base := 100 - float64(rank)
		for _, id := range x.C.SupportingEvidence {
			if selected[id] < base {
				selected[id] = base
			}
		}
		for _, id := range x.C.ContradictingEvidence {
			if selected[id] < base+5 {
				selected[id] = base + 5
			}
		}
		for _, id := range x.C.QualifyingEvidence {
			if selected[id] < base+2 {
				selected[id] = base + 2
			}
		}
	}
	eDocFreq := map[string]int{}
	eTokens := make([]map[string]int, len(rows))
	for i, r := range rows {
		t := tokenCounts(strings.Join([]string{r.Evidence.Assertion, r.Evidence.Subject, r.Evidence.Topic, strings.Join(r.Evidence.Modules, " "), strings.Join(r.Evidence.Keywords, " "), r.Source.InterviewTitle, r.Source.Focus}, " "))
		eTokens[i] = t
		for k := range t {
			eDocFreq[k]++
		}
	}
	var se []scoredEvidence
	for i, r := range rows {
		score := bm25Like(qTokens, eTokens[i], eDocFreq, len(rows))
		if v := selected[r.Evidence.EvidenceID]; v > score {
			score = v
		}
		if score > 0 {
			se = append(se, scoredEvidence{r, score})
		}
	}
	sort.Slice(se, func(i, j int) bool {
		if se[i].Score != se[j].Score {
			return se[i].Score > se[j].Score
		}
		return se[i].R.Evidence.EvidenceID < se[j].R.Evidence.EvidenceID
	})
	if len(se) > topE {
		se = se[:topE]
	}
	var b strings.Builder
	b.WriteString("# Case Query Packet\n\n")
	fmt.Fprintf(&b, "**Question:** %s\n\n", q)
	b.WriteString("Use this packet as a routing layer. Distinguish source-specific expert evidence from synthesis, surface competing claims, and do not infer beyond the evidence.\n\n")
	b.WriteString("## Relevant claim clusters\n\n")
	if len(sc) == 0 {
		b.WriteString("No normalized claim clusters matched. Review the evidence fallback below and state that the indexed evidence is limited.\n\n")
	} else {
		for _, x := range sc {
			c := x.C
			fmt.Fprintf(&b, "### %s — %s\n\n- **Claim:** %s\n- **Synthesis:** %s\n- **Module / topic / entity:** `%s` / `%s` / `%s`\n- **Confidence:** %s\n- **Supporting evidence:** %s\n- **Contradicting evidence:** %s\n- **Qualifying evidence:** %s\n", c.ClaimID, c.Status, c.NormalizedClaim, c.Synthesis, c.Module, c.Topic, c.PrimaryEntity, emptyAs(c.Confidence, "not stated"), joinOrNone(c.SupportingEvidence), joinOrNone(c.ContradictingEvidence), joinOrNone(c.QualifyingEvidence))
			if len(c.OpenQuestions) > 0 {
				fmt.Fprintf(&b, "- **Open questions:** %s\n", strings.Join(c.OpenQuestions, " | "))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("## Source-specific evidence\n\n")
	if len(se) == 0 {
		b.WriteString("No evidence records matched\n")
	} else {
		for _, x := range se {
			r := x.R
			e := r.Evidence
			fmt.Fprintf(&b, "### %s — %s\n\n- **Assertion:** %s\n- **Source:** %s | %s | `%s`\n- **Topic / modules:** `%s` / %s\n- **Type / basis / confidence:** %s / %s / %s\n- **Source anchors:** %s\n", e.EvidenceID, r.Source.CallID, e.Assertion, r.Source.InterviewTitle, r.Source.StakeholderType, r.Source.SourceFile, e.Topic, strings.Join(e.Modules, ", "), e.StatementType, e.Basis, e.Confidence, strings.Join(e.SourceAnchors, ", "))
			if len(e.Caveats) > 0 {
				fmt.Fprintf(&b, "- **Caveats:** %s\n", strings.Join(e.Caveats, " | "))
			}
			if e.Metric != nil {
				mb, _ := json.Marshal(e.Metric)
				fmt.Fprintf(&b, "- **Metric:** `%s`\n", string(mb))
			}
			fmt.Fprintf(&b, "- **Paths:** summary `%s`; extracted `%s`\n\n", r.Source.SummaryMarkdownPath, r.Source.ExtractedPath)
		}
	}
	b.WriteString("## Answering protocol\n\n1. Lead with the strongest evidence-backed conclusion\n2. Separate consensus, minority views, and true contradictions\n3. Explain scope differences such as period, geography, product, customer, scenario, and denominator\n4. Cite claim IDs and call/evidence IDs inline\n5. State confidence and evidence gaps\n6. Use `trace` only when exact source wording or surrounding context is needed\n")
	return b.String()
}

func emptyAs(s, d string) string {
	if strings.TrimSpace(s) == "" {
		return d
	}
	return s
}
func joinOrNone(xs []string) string {
	if len(xs) == 0 {
		return "none"
	}
	return strings.Join(xs, ", ")
}

func tokenize(s string) []string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	stop := map[string]bool{"the": true, "a": true, "an": true, "and": true, "or": true, "of": true, "to": true, "in": true, "on": true, "for": true, "from": true, "by": true, "is": true, "are": true, "was": true, "were": true, "what": true, "which": true, "how": true, "do": true, "does": true, "did": true, "we": true, "have": true, "has": true, "with": true, "about": true, "that": true, "this": true, "it": true}
	var out []string
	for _, w := range strings.Fields(b.String()) {
		if stop[w] || len(w) < 2 {
			continue
		}
		if strings.HasSuffix(w, "ies") && len(w) > 4 {
			w = w[:len(w)-3] + "y"
		} else if strings.HasSuffix(w, "s") && len(w) > 4 {
			w = w[:len(w)-1]
		}
		out = append(out, w)
	}
	return uniqueSorted(out)
}

func expandQueryTokens(tokens []string) []string {
	syn := map[string][]string{
		"supplier": {"vendor", "sourcing", "procurement"}, "vendor": {"supplier", "sourcing"}, "customer": {"buyer", "client", "account", "endcustomer"}, "competitor": {"rival", "peer", "competition"}, "price": {"pricing", "cost", "margin"}, "pricing": {"price", "cost", "margin"}, "capacity": {"utilization", "capex", "throughput"}, "qualification": {"onboarding", "qualified", "approval"}, "backlog": {"order", "reservation", "conversion"}, "delivery": {"leadtime", "otd", "schedule"}, "quality": {"rejection", "defect", "audit"}, "share": {"wallet", "marketshare", "sob"}, "contract": {"lta", "agreement", "mou", "takeorpay"},
	}
	out := append([]string{}, tokens...)
	for _, t := range tokens {
		out = append(out, syn[t]...)
	}
	return uniqueSorted(out)
}

func tokenCounts(s string) map[string]int {
	m := map[string]int{}
	for _, t := range tokenize(s) {
		m[t]++
	}
	return m
}
func bm25Like(q []string, doc map[string]int, df map[string]int, n int) float64 {
	score := 0.0
	for _, t := range q {
		tf := doc[t]
		if tf == 0 {
			continue
		}
		idf := math.Log(1 + float64(n+1)/float64(df[t]+1))
		score += idf * (1 + math.Log(float64(tf)))
	}
	return score
}

func cmdTrace(args []string) error {
	fs, rootFlagVal := rootFlag(args, "trace")
	evidenceArg := fs.String("evidence", "", "comma-separated evidence IDs")
	claimID := fs.String("claim", "", "claim ID")
	contextN := fs.Int("context", 1, "paragraphs of surrounding context")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	kdir := paths(root)["knowledge"]
	rows, err := readJSONL[EvidenceLedgerRow](filepath.Join(kdir, "evidence_ledger.jsonl"))
	if err != nil {
		return err
	}
	ids := []string{}
	if *evidenceArg != "" {
		for _, x := range strings.Split(*evidenceArg, ",") {
			if strings.TrimSpace(x) != "" {
				ids = append(ids, strings.TrimSpace(x))
			}
		}
	}
	if *claimID != "" {
		claims, err := readJSONL[FinalClaim](filepath.Join(kdir, "claim_registry.jsonl"))
		if err != nil {
			return err
		}
		found := false
		for _, c := range claims {
			if c.ClaimID == *claimID {
				ids = append(ids, c.SupportingEvidence...)
				ids = append(ids, c.ContradictingEvidence...)
				ids = append(ids, c.QualifyingEvidence...)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("claim not found: %s", *claimID)
		}
	}
	ids = uniqueSorted(ids)
	if len(ids) == 0 {
		return errors.New("provide --evidence or --claim")
	}
	byID := map[string]EvidenceLedgerRow{}
	for _, r := range rows {
		byID[r.Evidence.EvidenceID] = r
	}
	var b strings.Builder
	b.WriteString("# Source Trace\n\n")
	for _, id := range ids {
		r, ok := byID[id]
		if !ok {
			fmt.Fprintf(&b, "## %s\n\nEvidence ID not found\n\n", id)
			continue
		}
		fmt.Fprintf(&b, "## %s — %s\n\n- Source: `%s`\n- Assertion: %s\n- Anchors: %s\n\n", id, r.Source.CallID, r.Source.SourceFile, r.Evidence.Assertion, strings.Join(r.Evidence.SourceAnchors, ", "))
		paras, err := readAnchoredMarkdown(filepath.Join(root, filepath.FromSlash(r.Source.ExtractedPath)))
		if err != nil {
			return err
		}
		selected := map[int]bool{}
		for _, a := range r.Evidence.SourceAnchors {
			m := anchorRe.FindStringSubmatch(a)
			if m == nil {
				continue
			}
			lo, _ := strconv.Atoi(m[1])
			hi := lo
			if m[2] != "" {
				hi, _ = strconv.Atoi(m[2])
			}
			lo -= *contextN
			if lo < 1 {
				lo = 1
			}
			hi += *contextN
			for n := lo; n <= hi; n++ {
				selected[n] = true
			}
		}
		var nums []int
		for n := range selected {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		for _, n := range nums {
			if txt, ok := paras[n]; ok {
				fmt.Fprintf(&b, "- **P%04d:** %s\n", n, txt)
			}
		}
		b.WriteString("\n")
	}
	fmt.Print(b.String())
	return nil
}

func readAnchoredMarkdown(path string) (map[int]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[int]string{}
	scanner := bufio.NewScanner(f)
	// Transcripts can contain long paragraphs; raise the scanner limit well beyond the default.
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	current := 0
	var body strings.Builder
	flush := func() {
		if current > 0 {
			out[current] = strings.TrimSpace(body.String())
		}
		body.Reset()
	}
	anchorLine := regexp.MustCompile(`^\[P(\d{4})\]\s*(.*)$`)
	for scanner.Scan() {
		line := scanner.Text()
		if m := anchorLine.FindStringSubmatch(line); m != nil {
			flush()
			current, _ = strconv.Atoi(m[1])
			body.WriteString(m[2])
			continue
		}
		if current > 0 {
			if body.Len() > 0 {
				body.WriteByte('\n')
			}
			body.WriteString(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	flush()
	return out, nil
}

func cmdStatus(args []string) error {
	fs, rootFlagVal := rootFlag(args, "status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := absRoot(*rootFlagVal)
	if err != nil {
		return err
	}
	reg, err := loadRegistry(root)
	if err != nil {
		return err
	}
	counts := map[string]int{}
	for _, c := range reg.Calls {
		counts[c.Status]++
		if c.SourceMissing {
			counts["source_missing"]++
		}
	}
	jobs := map[string]int{}
	var m CompileJobManifest
	if err := readJSON(paths(root)["manifest"], &m); err == nil {
		for _, j := range m.Jobs {
			jobs[j.Status]++
		}
	}
	claims := 0
	if cs, err := readJSONL[FinalClaim](filepath.Join(paths(root)["knowledge"], "claim_registry.jsonl")); err == nil {
		claims = len(cs)
	}
	evidence := 0
	if es, err := readJSONL[EvidenceLedgerRow](filepath.Join(paths(root)["knowledge"], "evidence_ledger.jsonl")); err == nil {
		evidence = len(es)
	}
	return printJSON(map[string]interface{}{"root": root, "calls": counts, "registered_calls": len(reg.Calls), "compile_jobs": jobs, "claims": claims, "evidence": evidence, "master_docx": relPath(root, filepath.Join(paths(root)["compiled"], defaultConfig().MasterFilename)), "case_map": relPath(root, filepath.Join(paths(root)["knowledge"], "CASE_MAP.md"))})
}

func cmdSelfTest(args []string) error {
	fs := flag.NewFlagSet("self-test", flag.ContinueOnError)
	keep := fs.Bool("keep", false, "keep temporary directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "expertcall-selftest-")
	if err != nil {
		return err
	}
	if !*keep {
		defer os.RemoveAll(tmp)
	}
	if err := ensureDirs(tmp); err != nil {
		return err
	}

	spec := SummarySpec{
		SchemaVersion: "1.0",
		CallID:        "EC-0001",
		SourceFile:    "Synthetic.docx",
		SourceSHA256:  strings.Repeat("a", 64),
		Interviews: []Interview{
			{
				Title: "Interview — Procurement Director, Example OEM",
				Framing: []string{
					"Stakeholder type: OEM-side sourcing insider",
					"Focus: Supplier capacity, qualification, and pricing mechanics",
				},
				Themes: []Theme{
					{
						Lead: "Capacity is the binding constraint",
						Points: []string{
							"The expert estimates that qualified suppliers are operating above 90% utilization, making delivery more important than marginal price differences.",
							"He distinguishes committed capacity from announced capex, noting that new lines require customer qualification before they become usable supply.",
						},
					},
					{
						Lead: "Qualification is staged rather than binary",
						Points: []string{
							"Initial part approval may take 12–18 months, while full supplier maturity can take several years and depends on repeat production performance.",
						},
					},
				},
			},
		},
	}
	out := filepath.Join(tmp, "test.docx")
	if err := writeSummaryDOCX(out, spec.Interviews, false); err != nil {
		return err
	}
	if inf, err := os.Stat(out); err != nil || inf.Size() < 1000 {
		return errors.New("DOCX writer self-test failed")
	}
	zr, err := zip.OpenReader(out)
	if err != nil {
		return err
	}
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	zr.Close()
	for _, n := range []string{"[Content_Types].xml", "word/document.xml", "word/styles.xml", "word/numbering.xml"} {
		if !names[n] {
			return fmt.Errorf("DOCX missing %s", n)
		}
	}
	tempDir := "removed"
	if *keep {
		tempDir = tmp
	}
	return printJSON(map[string]interface{}{
		"status":       "self_test_passed",
		"tool_version": toolVersion,
		"temp_dir":     tempDir,
	})
}
func printJSON(v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// DOCX writer
func writeSummaryDOCX(path string, interviews []Interview, master bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	_ = os.Remove(tmp)
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(f)
	files := map[string]string{
		"[Content_Types].xml":          contentTypesXML(),
		"_rels/.rels":                  rootRelsXML(),
		"docProps/core.xml":            corePropsXML(path),
		"docProps/app.xml":             appPropsXML(),
		"word/document.xml":            documentXML(interviews),
		"word/_rels/document.xml.rels": documentRelsXML(),
		"word/styles.xml":              stylesXML(),
		"word/numbering.xml":           numberingXML(),
		"word/settings.xml":            settingsXML(),
		"word/fontTable.xml":           fontTableXML(),
		"word/theme/theme1.xml":        themeXML(),
	}
	order := []string{"[Content_Types].xml", "_rels/.rels", "docProps/core.xml", "docProps/app.xml", "word/document.xml", "word/_rels/document.xml.rels", "word/styles.xml", "word/numbering.xml", "word/settings.xml", "word/fontTable.xml", "word/theme/theme1.xml"}
	for _, name := range order {
		w, err := zw.Create(name)
		if err != nil {
			zw.Close()
			f.Close()
			return err
		}
		if _, err := io.WriteString(w, files[name]); err != nil {
			zw.Close()
			f.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}
	return os.Rename(tmp, path)
}

func xmlEsc(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
func textRun(text string, bold bool) string {
	pr := ""
	if bold {
		pr = "<w:rPr><w:b/><w:bCs/></w:rPr>"
	}
	space := ""
	if strings.HasPrefix(text, " ") || strings.HasSuffix(text, " ") {
		space = ` xml:space="preserve"`
	}
	return `<w:r>` + pr + `<w:t` + space + `>` + xmlEsc(text) + `</w:t></w:r>`
}
func paraHeading(text string, breakBefore bool) string {
	bb := ""
	if breakBefore {
		bb = "<w:pageBreakBefore/>"
	}
	return `<w:p><w:pPr><w:pStyle w:val="Heading2"/><w:keepNext/>` + bb + `<w:spacing w:before="400" w:after="200"/></w:pPr>` + textRun(text, true) + `</w:p>`
}
func paraBullet(runs string, numID int, after int) string {
	return fmt.Sprintf(`<w:p><w:pPr><w:numPr><w:ilvl w:val="0"/><w:numId w:val="%d"/></w:numPr><w:spacing w:after="%d"/><w:widowControl/></w:pPr>%s</w:p>`, numID, after, runs)
}
func framingRuns(s string) string {
	idx := strings.Index(s, ":")
	if idx >= 0 {
		return textRun(s[:idx+1], true) + textRun(s[idx+1:], false)
	}
	return textRun(s, false)
}
func documentXML(interviews []Interview) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><w:body>`)
	for i, in := range interviews {
		b.WriteString(paraHeading(in.Title, i > 0))
		for _, f := range in.Framing {
			b.WriteString(paraBullet(framingRuns(f), 1, 160))
		}
		for _, t := range in.Themes {
			b.WriteString(paraBullet(textRun(t.Lead, true), 1, 80))
			for _, p := range t.Points {
				b.WriteString(paraBullet(textRun(p, false), 2, 120))
			}
		}
	}
	b.WriteString(`<w:sectPr><w:pgSz w:w="12240" w:h="15840"/><w:pgMar w:top="1080" w:right="1080" w:bottom="1080" w:left="1080" w:header="720" w:footer="720" w:gutter="0"/><w:cols w:space="720"/><w:docGrid w:linePitch="360"/></w:sectPr></w:body></w:document>`)
	return b.String()
}
func contentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/><Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/><Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/><Override PartName="/word/fontTable.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.fontTable+xml"/><Override PartName="/word/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`
}
func rootRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>`
}
func documentRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings" Target="settings.xml"/><Relationship Id="rId4" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/fontTable" Target="fontTable.xml"/><Relationship Id="rId5" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="theme/theme1.xml"/></Relationships>`
}
func corePropsXML(path string) string {
	t := nowUTC()
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>` + xmlEsc(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))) + `</dc:title><dc:creator>Expert Call Intelligence</dc:creator><cp:lastModifiedBy>Expert Call Intelligence</cp:lastModifiedBy><dcterms:created xsi:type="dcterms:W3CDTF">` + t + `</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">` + t + `</dcterms:modified></cp:coreProperties>`
}
func appPropsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><Application>Expert Call Intelligence</Application><AppVersion>1.0</AppVersion></Properties>`
}
func stylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Aptos" w:hAnsi="Aptos" w:eastAsia="Aptos" w:cs="Aptos"/><w:sz w:val="22"/><w:szCs w:val="22"/><w:lang w:val="en-US"/></w:rPr></w:rPrDefault><w:pPrDefault><w:pPr><w:spacing w:after="120" w:line="276" w:lineRule="auto"/></w:pPr></w:pPrDefault></w:docDefaults><w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/><w:qFormat/></w:style><w:style w:type="paragraph" w:styleId="Heading2"><w:name w:val="heading 2"/><w:basedOn w:val="Normal"/><w:next w:val="Normal"/><w:uiPriority w:val="9"/><w:qFormat/><w:pPr><w:keepNext/><w:keepLines/><w:outlineLvl w:val="1"/></w:pPr><w:rPr><w:rFonts w:ascii="Aptos Display" w:hAnsi="Aptos Display"/><w:b/><w:bCs/><w:sz w:val="30"/><w:szCs w:val="30"/><w:color w:val="1F1F1F"/></w:rPr></w:style></w:styles>`
}
func numberingXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:abstractNum w:abstractNumId="0"><w:multiLevelType w:val="singleLevel"/><w:lvl w:ilvl="0"><w:start w:val="1"/><w:numFmt w:val="bullet"/><w:lvlText w:val="•"/><w:lvlJc w:val="left"/><w:pPr><w:tabs><w:tab w:val="num" w:pos="360"/></w:tabs><w:ind w:left="360" w:hanging="260"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/></w:rPr></w:lvl></w:abstractNum><w:abstractNum w:abstractNumId="1"><w:multiLevelType w:val="singleLevel"/><w:lvl w:ilvl="0"><w:start w:val="1"/><w:numFmt w:val="bullet"/><w:lvlText w:val="•"/><w:lvlJc w:val="left"/><w:pPr><w:tabs><w:tab w:val="num" w:pos="720"/></w:tabs><w:ind w:left="720" w:hanging="260"/></w:pPr><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/></w:rPr></w:lvl></w:abstractNum><w:num w:numId="1"><w:abstractNumId w:val="0"/></w:num><w:num w:numId="2"><w:abstractNumId w:val="1"/></w:num></w:numbering>`
}
func settingsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:zoom w:percent="100"/><w:defaultTabStop w:val="720"/><w:characterSpacingControl w:val="doNotCompress"/><w:compat><w:compatSetting w:name="compatibilityMode" w:uri="http://schemas.microsoft.com/office/word" w:val="15"/></w:compat></w:settings>`
}
func fontTableXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><w:fonts xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:font w:name="Aptos"/><w:font w:name="Aptos Display"/><w:font w:name="Arial"/></w:fonts>`
}
func themeXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><a:theme xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" name="Office"><a:themeElements><a:clrScheme name="Office"><a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1><a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1><a:dk2><a:srgbClr val="44546A"/></a:dk2><a:lt2><a:srgbClr val="E7E6E6"/></a:lt2><a:accent1><a:srgbClr val="4472C4"/></a:accent1><a:accent2><a:srgbClr val="ED7D31"/></a:accent2><a:accent3><a:srgbClr val="A5A5A5"/></a:accent3><a:accent4><a:srgbClr val="FFC000"/></a:accent4><a:accent5><a:srgbClr val="5B9BD5"/></a:accent5><a:accent6><a:srgbClr val="70AD47"/></a:accent6><a:hlink><a:srgbClr val="0563C1"/></a:hlink><a:folHlink><a:srgbClr val="954F72"/></a:folHlink></a:clrScheme><a:fontScheme name="Office"><a:majorFont><a:latin typeface="Aptos Display"/></a:majorFont><a:minorFont><a:latin typeface="Aptos"/></a:minorFont></a:fontScheme><a:fmtScheme name="Office"><a:fillStyleLst/><a:lnStyleLst/><a:effectStyleLst/><a:bgFillStyleLst/></a:fmtScheme></a:themeElements></a:theme>`
}
