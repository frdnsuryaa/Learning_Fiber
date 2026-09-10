CREATE TABLE IF NOT EXISTS nilai (
    idnilai SERIAL PRIMARY KEY,
    namamatkul VARCHAR(100) NOT NULL,
    nilai VARCHAR(10) NOT NULL,
    idstudent VARCHAR(36) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_nilai_student FOREIGN KEY (idstudent) REFERENCES students(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_nilai_idstudent ON nilai (idstudent);
