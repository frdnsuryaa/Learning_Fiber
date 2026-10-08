package helper

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func TestParseStudentQuery_ValidValues(t *testing.T) {
	app := fiber.New()
	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod(fasthttp.MethodGet)
	ctx.Request.SetRequestURI("/api/v1/students?page=2&per_page=20&search=Rina&prodi=Sistem%20Informasi&angkatan=2023&sort=-ipk_terakhir")
	c := app.AcquireCtx(ctx)
	defer app.ReleaseCtx(c)
	q := ParseStudentQuery(c)

	if q.Page != 2 || q.PerPage != 20 || q.Search != "Rina" || q.Prodi != "Sistem Informasi" || q.Sort != "-ipk_terakhir" {
		t.Fatalf("unexpected query parse: %#v", q)
	}
	if q.Angkatan == nil || *q.Angkatan != 2023 {
		t.Fatal("angkatan should be parsed")
	}
}

func TestParseStudentQuery_DefaultsAndUnsupportedSort(t *testing.T) {
	app := fiber.New()
	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod(fasthttp.MethodGet)
	ctx.Request.SetRequestURI("/api/v1/students?sort=unknown")
	c := app.AcquireCtx(ctx)
	defer app.ReleaseCtx(c)
	q := ParseStudentQuery(c)

	if q.Page != 1 || q.PerPage != 10 || q.Sort != "nama" {
		t.Fatalf("unexpected defaults: %#v", q)
	}
}
