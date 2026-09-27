package moderation

func splitContentModerationText(text string, chunkSize int, overlap int) []string {
	chunks := make([]string, 0, countContentModerationTextChunks(text, chunkSize, overlap))
	forEachContentModerationTextBatch(text, chunkSize, overlap, contentModerationTextBatchSize, func(_ int, batch []string) {
		chunks = append(chunks, batch...)
	})
	return chunks
}
