CREATE TABLE IF NOT EXISTS files (
                                     id UUID PRIMARY KEY,
                                     name TEXT NOT NULL,
                                     path TEXT NOT NULL,
                                     media_type TEXT NOT NULL,
                                     created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );