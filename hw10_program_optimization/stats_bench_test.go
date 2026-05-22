package hw10programoptimization

import (
	"archive/zip"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	r, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatal(err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		b.Fatalf("expected 1 file in archive, got %d", len(r.File))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		data, err := r.File[0].Open()
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		_, err = GetDomainStat(data, "biz")
		if err != nil {
			b.Fatal(err)
		}

		b.StopTimer()
		data.Close()
		b.StartTimer()
	}
}
