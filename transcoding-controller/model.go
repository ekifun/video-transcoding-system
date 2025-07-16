package main

type TranscodeRequest struct {
    StreamName  string   `json:"stream_name"`
    InputURL    string   `json:"input_url"`
    Resolutions []string `json:"resolutions"`
    Codec       string   `json:"codec"`
    GopSize     int      `json:"gop_size"`      // ✅ Added field
    KeyintMin   int      `json:"keyint_min"`    // ✅ Added field
}

type TranscodeJob struct {
    JobID          string `json:"job_id"`
    InputURL       string `json:"input_url"`
    Representation string `json:"representation"`
    Resolution     string `json:"resolution"`   // e.g., 1280x720
    Bitrate        string `json:"bitrate"`      // e.g., 2500k
    Codec          string `json:"codec"`        // e.g., h264
    OutputPath     string `json:"output_path"`
    GopSize        int    `json:"gop_size"`     // ✅ Added field
    KeyintMin      int    `json:"keyint_min"`   // ✅ Added field
}
