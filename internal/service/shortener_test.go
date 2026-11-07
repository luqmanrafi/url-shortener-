package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/luqmanrafi/urlshortener/internal/model"
)

type fakeRepo struct {
	data    map[string]*model.URL
	pingErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{data: map[string]*model.URL{}}
}

func (f *fakeRepo) Save(ctx context.Context, code, originalURL string, expiresAt *time.Time)error{
	if _, ada := f.data[code]; ada{
		return model.ErrCodeExists
	}
	f.data[code] = &model.URL{
		Code: code,
		Original: originalURL,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (f *fakeRepo) FindByCode(ctx context.Context, code string)(*model.URL, error){
	u, ada := f.data[code]
	if !ada{
		return nil, model.ErrNotFound
	}
	return u, nil
}

func (f *fakeRepo) IncrementClick(ctx context.Context, code string)error{
	if u, ada := f.data[code]; ada{
		u.ClickCount++
	}
	return nil

}

func (f *fakeRepo) GetStats(ctx context.Context, code string) (*model.URL, error) {
	return f.FindByCode(ctx, code)
}

func (f *fakeRepo) Ping(ctx context.Context) error {
	return f.pingErr
}

func TestShorten_ValidURL(t *testing.T){
	svc := New(newFakeRepo())
	code, err := svc.Shorten(context.Background(), "https://go.dev", "", 0)
	if err != nil{
		t.Fatalf("tidak boleh error, dapat: %v", err)
	}
	if len(code)!=7{
		t.Fatalf("panjang code harus 7, dapat: %v", len(code))
	}
}
func TestShorten_InvalidURL(t *testing.T){
	svc := New(newFakeRepo())

	_, err:= svc.Shorten(context.Background(), "javascript:alert(1)", "", 0)
	if !errors.Is(err, model.ErrInvalidUrl){
		t.Fatalf("error harus InvalidURL, dapat %v", err)
	}
}

func TestReady(t *testing.T){
	repo:= newFakeRepo()
	svc:= New(repo)

	if err := svc.Ready(context.Background()); err != nil{
		t.Fatalf("harus sehat, dapat: %v", err)
	}
	
	repo.pingErr = errors.New("db mati")
	if  err:= svc.Ready(context.Background()); err == nil{
		t.Fatalf("harus mendapatkan error database mati")
	}
}

func TestCustomAlias(t *testing.T){
	svc := New(newFakeRepo())

	code, err := svc.Shorten(context.Background(), "https://go.dev", "belajar-go", 0)

	if err != nil{
		t.Fatalf("tidak boleh error, dapat : %v", err)
	}
	if code != "belajar-go"{
		t.Errorf("kode harus belajar-go, dapat %q", err)
	}
}

func TestShorten_DuplicateAlias(t *testing.T){
	svc := New(newFakeRepo())

	_, err := svc.Shorten(context.Background(), "https://go.dev", "belajar-go", 0)
	if err != nil {
		t.Fatalf("pemakaian pertama harus sukses, dapat: %v", err)
	}

	_, err = svc.Shorten(context.Background(), "https://example.com", "belajar-go", 0)
	if !errors.Is(err, model.ErrCodeExists) {
		t.Errorf("harus ErrCodeExists, dapat: %v", err)
	}
}
func TestResolve_Expired(t *testing.T) {
	repo := newFakeRepo()
	svc := New(repo)
	sekarang := time.Now()
	svc.now = func() time.Time { return sekarang }

	code, _ := svc.Shorten(context.Background(), "https://go.dev", "", time.Hour)

	// majukan jam 2 jam
	svc.now = func() time.Time { return sekarang.Add(2 * time.Hour) }

	_, err := svc.Resolve(context.Background(), code)
	if !errors.Is(err, model.ErrExpired) {
		t.Errorf("harus ErrExpired, dapat: %v", err)
	}
	if repo.data[code].ClickCount != 0 {
		t.Errorf("klik pada link kadaluarsa tidak boleh dihitung, dapat %d", repo.data[code].ClickCount)
	}
}
func TestValidateAlias(t *testing.T) {
	tests := []struct {
		alias string
		valid bool
	}{
		{"belajar-go", true},
		{"ab", false},                 // terlalu pendek
		{"a b c", false},              // spasi
		{"../etc", false},             // karakter terlarang
		{"API", false},                // reserved, beda huruf
		{"healthz", false},            // reserved
		{strings.Repeat("a", 31), false}, // terlalu panjang
	}
	for _, tt := range tests {
		err := validateAlias(tt.alias)
		if tt.valid && err != nil {
			t.Errorf("%q harus valid, dapat: %v", tt.alias, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("%q harus ditolak", tt.alias)
		}
	}
}

func TestResolve_NotYetExpired(t *testing.T){
	svc := New(newFakeRepo())
	sekarang := time.Now()
	svc.now = func() time.Time {return sekarang}

	code, err := svc.Shorten(context.Background(), "https://go.dev", "", time.Hour)
	if err != nil{
		t.Fatalf("shorten gagal: %v", err)
	}

	svc.now = func() time.Time {return sekarang.Add(30 * time.Minute)}
	got, err := svc.Resolve(context.Background(), code)
	if err != nil {
		t.Fatalf("tidak boleh error, dapat: %v", err)
	}
	if got != "https://go.dev"{
		t.Errorf("url salah: %q", got)
	}
}

func TestResolve_NoExpiry(t *testing.T){
	svc := New(newFakeRepo())
	sekarang := time.Now()
	svc.now = func() time.Time { return sekarang}
	code, err := svc.Shorten(context.Background(), "https://go.dev", "", 0)
	if err != nil {
		t.Fatalf("shorten gagal: %v", err)
	}

	svc.now= func() time.Time {return sekarang.Add(10 * 365 * 24 * time.Hour)}
	if _, err := svc.Resolve(context.Background(), code); err != nil {
		t.Errorf("tanpa expiry tidak boleh error, dapat : %v", err)
	}
}