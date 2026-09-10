package model

import "time"

type Nilai struct {
	IDNilai    int       `json:"idnilai"`
	NamaMatkul string    `json:"namamatkul"`
	Nilai      string    `json:"nilai"`
	IDStudent  string    `json:"idstudent"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
}
