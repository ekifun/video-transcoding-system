package main

type TranscodeRequest struct {
	InputURL string   `json:"input_url"`
	Codec    string   `json:"codec"` // e.g., h264
	Resolutions []string `json:"resolutions"` // e.g., ["144p", "360p", "720p"]
}

type TranscodeJob struct {
	JobID        string `json:"job_id"`
	InputURL     string `json:"input_url"`
	Representation string `json:"representation"`
	Resolution   string `json:"resolution"`   // e.g., 1280x720
	Bitrate      string `json:"bitrate"`      // e.g., 2500k
	Codec        string `json:"codec"`        // e.g., h264
	OutputPath   string `json:"output_path"`
}
