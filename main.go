package main

import "fmt"

func main() {

	// DEKLARASI 5 VARIABEL DENGAN TIPE DATA BERBEDA

	var namaProgram string = "Sistem Manajemen Akademik"
	var totalKelas int = 2
	var kriteriaLulus float64 = 75.0
	var isSistemAktif bool = true
	daftarMataKuliah := []string{"UI/UX,", "BackEnd"}

	fmt.Println("INFORMASI SISTEM AKADEMIK")
	fmt.Printf("Program        : %s\n", namaProgram)
	fmt.Printf("Total Kelas    : %d\n", totalKelas)
	fmt.Printf("Kriteria Lulus : %.1f\n", kriteriaLulus)
	fmt.Printf("Status Aktif   : %t\n", isSistemAktif)
	fmt.Printf("Mata Kuliah    : %v\n", daftarMataKuliah)

	// DEKLARASI DAN OPERASI PADA MAP MAHASISWA

	dataMahasiswa := make(map[string]float64)

	// Menambah Data
	dataMahasiswa["Alan"] = 85.4
	dataMahasiswa["Fahmi"] = 92.8
	dataMahasiswa["Rizky"] = 68.1
	dataMahasiswa["Bagus"] = 78.0

	fmt.Println("\n OPERASI MAP MAHASISWA")

	// Membaca Data dengan Pengecekan Keberadaan
	namaCari := "Alan"
	nilaiAlan, ada := dataMahasiswa[namaCari]
	if ada {
		fmt.Printf("Data ditemukan: %s mendapat nilai %.1f\n", namaCari, nilaiAlan)
	} else {
		fmt.Printf("Data %s tidak ditemukan.\n", namaCari)
	}

	// Pengecekan data yang tidak ada di map
	namaCari2 := "Eko"
	if nilaiEko, adaEko := dataMahasiswa[namaCari2]; adaEko {
		fmt.Printf("Data %s mendapat nilai %.1f\n", namaCari2, nilaiEko)
	} else {
		fmt.Printf("Data %s tidak ada di dalam map.\n", namaCari2)
	}

	// Menghapus Data dari Map
	fmt.Println("\nMenghapus data mahasiswa bernama Alan")
	delete(dataMahasiswa, "Alan")

	// Menelusuri Seluruh Isi Map
	fmt.Println("\nDaftar Seluruh Mahasiswa Tersisa:")
	for nama, nilai := range dataMahasiswa {
		fmt.Printf("Nama: %-5s  Nilai: %.1f\n", nama, nilai)
	}
}