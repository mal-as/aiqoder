package models

type Chunk struct {
	RepoID    string
	FilePath  string
	Content   string
	Embedding []float32
}

type ChunkResult struct {
	FilePath string
	Content  string
}
