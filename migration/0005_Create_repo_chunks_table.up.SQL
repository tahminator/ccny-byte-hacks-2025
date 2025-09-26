CREATE TABLE repo_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_hash TEXT NOT NULL,
    source TEXT NOT NULL,
    chunk TEXT NOT NULL,
    embedding vector(768),
    -- Merge conflict specific fields
    file_type TEXT, -- 'conflict', 'context', 'resolved'
    conflict_section TEXT, -- Which part of the conflict this chunk represents
    line_start INTEGER,
    line_end INTEGER,
    chunk_type TEXT DEFAULT 'code', -- 'code', 'comment', 'import', 'function'
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index for efficient similarity search
CREATE INDEX ON repo_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Create indexes for merge conflict analysis
CREATE INDEX idx_repo_chunks_repo_hash ON repo_chunks(repo_hash);
CREATE INDEX idx_repo_chunks_file_type ON repo_chunks(file_type);
CREATE INDEX idx_repo_chunks_source ON repo_chunks(source);
CREATE INDEX idx_repo_chunks_chunk_type ON repo_chunks(chunk_type);
