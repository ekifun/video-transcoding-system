type TranscodeJob struct {
	JobID          string `json:"job_id"`
	InputURL       string `json:"input_url"`
	Representation string `json:"representation"`
	Resolution     string `json:"resolution"`
	Bitrate        string `json:"bitrate"`
	Codec          string `json:"codec"`
	OutputPath     string `json:"output_path"`
}
