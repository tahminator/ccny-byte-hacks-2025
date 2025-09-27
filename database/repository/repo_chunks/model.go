package repo_chunks

type RepoChunk struct {
	RepoHash        string    `db:"repo_hash" json:"repo_hash"`
	Source          string    `db:"source" json:"source"`
	Chunk           string    `db:"chunk" json:"chunk"`
	Embedding       []float64 `db:"embedding" json:"embedding"`
	FileType        string    `db:"file_type" json:"file_type"`
	ConflictSection string    `db:"conflict_section" json:"conflict_section"`
	LineStart       int       `db:"line_start" json:"line_start"`
	LineEnd         int       `db:"line_end" json:"line_end"`
	ChunkType       string    `db:"chunk_type" json:"chunk_type"`
}

type SimilarChunk struct {
	Source          string  `db:"source" json:"source"`
	Chunk           string  `db:"chunk" json:"chunk"`
	Distance        float64 `db:"distance" json:"distance"`
	FileType        string  `db:"file_type" json:"file_type"`
	ConflictSection string  `db:"conflict_section" json:"conflict_section"`
	LineStart       int     `db:"line_start" json:"line_start"`
	LineEnd         int     `db:"line_end" json:"line_end"`
	ChunkType       string  `db:"chunk_type" json:"chunk_type"`
}
