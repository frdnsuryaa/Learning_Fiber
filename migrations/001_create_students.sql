CREATE TABLE IF NOT EXISTS students (
    id VARCHAR(36) PRIMARY KEY,
    nim VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    grade DOUBLE PRECISION NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT students_nim_unique UNIQUE (nim)
);


CREATE INDEX IF NOT EXISTS idx_students_name_lower
    ON students (LOWER(name));


CREATE INDEX IF NOT EXISTS idx_students_created_at
    ON students (created_at DESC);
