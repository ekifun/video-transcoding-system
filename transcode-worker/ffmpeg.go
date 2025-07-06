func HandleTranscodeJob(job TranscodeJob) {
	localInput, err := DownloadInput(job.InputURL, job.JobID)
	if err != nil {
		log.Printf("❌ Download failed: %v", err)
		return
	}

	outputPath := fmt.Sprintf("/tmp/output_%s.mp4", job.Representation)
	cmd := exec.Command("ffmpeg", "-i", localInput, "-vf", fmt.Sprintf("scale=%s", job.Resolution), 
		"-b:v", job.Bitrate, "-c:v", job.Codec, "-y", outputPath)

	err = cmd.Run()
	if err != nil {
		log.Printf("❌ FFmpeg failed: %v", err)
		return
	}

	log.Printf("✅ Transcoding done: %s", outputPath)
	PublishStatus(job.JobID, job.Representation, "done") // publish to transcode-status
}
