package main

// TranscodeJob represents a single video transcoding task
type TranscodeJob struct {
	JobID          string `json:"job_id"`          // Unique identifier for the job
	InputURL       string `json:"input_url"`       // Source video URL
	Representation string `json:"representation"`  // e.g., 720p
	Resolution     string `json:"resolution"`      // e.g., 1280x720
	Bitrate        string `json:"bitrate"`         // e.g., 2500k
	Codec 		   string `json:"codec"`           // Codec to use (e.g., h264, hevc, vvc, vp9)	OutputPath     string `json:"output_path"`     // Path to save the transcoded video
	GopSize        int    `json:"gop_size"`        // Group of Pictures (GOP) size, e.g., 48
	KeyintMin      int    `json:"keyint_min"`      // Minimum interval between keyframes, e.g., 48
}
