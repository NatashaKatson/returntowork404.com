package handlers

import (
	"encoding/json"
	"net/http"
	"slices"

	"whatdidimiss/cache"
	"whatdidimiss/claude"
)

// Valid industries - add more as needed
var validIndustries = []string{
	"software-development",
	"marketing",
	"healthcare",
	"legal",
}

// Valid time periods
var validTimePeriods = []string{
	"6-months",
	"1-year",
	"2-3-years",
	"5-years",
	"10-years",
}

// Human-readable labels for display
var industryLabels = map[string]string{
	"software-development": "Software Development",
	"marketing":            "Marketing",
	"healthcare":           "Healthcare",
	"legal":                "Legal",
}

var timePeriodLabels = map[string]string{
	"6-months":   "6 months",
	"1-year":     "1 year",
	"2-3-years":  "2-3 years",
	"5-years":    "5+ years",
	"10-years":   "10+ years",
}

type APIHandler struct {
	cache  *cache.MemoryCache
	claude *claude.Client
}

type CatchUpRequest struct {
	Industry   string `json:"industry"`
	TimePeriod string `json:"time_period"`
}

type CatchUpResponse struct {
	Summary  string `json:"summary"`
	Industry string `json:"industry"`
	Period   string `json:"period"`
	Cached   bool   `json:"cached"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func NewAPIHandler(cache *cache.MemoryCache, claude *claude.Client) *APIHandler {
	return &APIHandler{
		cache:  cache,
		claude: claude,
	}
}

func (h *APIHandler) CatchUp(w http.ResponseWriter, r *http.Request) {
	var req CatchUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Validate industry
	if !slices.Contains(validIndustries, req.Industry) {
		writeError(w, http.StatusBadRequest, "Invalid industry", "Must be one of: software-development, marketing, healthcare, legal")
		return
	}

	// Validate time period
	if !slices.Contains(validTimePeriods, req.TimePeriod) {
		writeError(w, http.StatusBadRequest, "Invalid time period", "Must be one of: 6-months, 1-year, 2-3-years, 5-years, 10-years")
		return
	}

	// Check cache first
	cacheKey := req.Industry + ":" + req.TimePeriod
	if cached, err := h.cache.Get(r.Context(), cacheKey); err == nil && cached != "" {
		writeJSON(w, http.StatusOK, CatchUpResponse{
			Summary:  cached,
			Industry: industryLabels[req.Industry],
			Period:   timePeriodLabels[req.TimePeriod],
			Cached:   true,
		})
		return
	}

	// Generate summary from Claude
	summary, err := h.claude.GenerateSummary(r.Context(), industryLabels[req.Industry], timePeriodLabels[req.TimePeriod])
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate summary", err.Error())
		return
	}

	// Cache the response (7 days)
	if err := h.cache.Set(r.Context(), cacheKey, summary); err != nil {
		// Log but don't fail - caching is best-effort
		// log.Printf("Failed to cache response: %v", err)
	}

	writeJSON(w, http.StatusOK, CatchUpResponse{
		Summary:  summary,
		Industry: industryLabels[req.Industry],
		Period:   timePeriodLabels[req.TimePeriod],
		Cached:   false,
	})
}

func (h *APIHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, error string, details string) {
	writeJSON(w, status, ErrorResponse{Error: error, Details: details})
}
