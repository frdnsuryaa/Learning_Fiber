package main

import "fmt"

// Student merepresentasikan data mahasiswa
type Student struct {
	ID       string
	Name     string
	Grade    float64
	IsActive bool
}

// GetInfo: Value Receiver (Hanya membaca data / Read-only)
func (s Student) GetInfo() string {
	status := "Tidak Aktif"
	if s.IsActive {
		status = "Aktif"
	}
	return fmt.Sprintf("ID: %s | Nama: %s | Nilai: %.2f | Status: %s", s.ID, s.Name, s.Grade, status)
}

// UpdateGrade: Pointer Receiver (mengubah nilai Grade asli)
func (s *Student) UpdateGrade(grade float64) {
	s.Grade = grade
}

// Activate: Pointer Receiver (mengubah IsActive asli menjadi true)
func (s *Student) Activate() {
	s.IsActive = true
}

// Deactivate: Pointer Receiver (mengubah IsActive asli menjadi false)
func (s *Student) Deactivate() {
	s.IsActive = false
}

func main() {

	mhs := Student{
		ID:       "MHS001",
		Name:     "Budi Santoso",
		Grade:    78.5,
		IsActive: false,
	}

	fmt.Println("KONDISI AWAL")
	fmt.Println(mhs.GetInfo())

	fmt.Println("SETELAH AKTIVASI DAN UPDATE NILAI")
	mhs.Activate()
	mhs.UpdateGrade(90.0)
	fmt.Println(mhs.GetInfo())

	fmt.Println("SETELAH DEAKTIVASI")
	mhs.Deactivate()
	fmt.Println(mhs.GetInfo())
}
