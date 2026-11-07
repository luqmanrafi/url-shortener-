package service

import (
	"context"
	"errors"
	"testing"

	"github.com/luqmanrafi/urlshortener/internal/model"
)

type fakeRepo struct {
	data    map[string]string
	pingErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{data: map[string]string{}}
}

func (f *fakeRepo) Save(ctx context.Context, code, url string)error{
	f.data[code] = url
	return nil
}

func (f *fakeRepo) FindByCode(ctx context.Context, code string)(string, error){
	url, ada := f.data[code]
	if !ada{
		return "", model.ErrNotFound
	}
	return url, nil
}

func (f *fakeRepo) IncrementClick(ctx context.Context, code string)error{
	return nil
}

func (f *fakeRepo) GetStats(ctx context.Context, code string) (*model.URL, error) {
	return nil, model.ErrNotFound
}

func (f *fakeRepo) Ping(ctx context.Context) error {
	return f.pingErr
}

func TestShorten_ValidURL(t *testing.T){
	svc := New(newFakeRepo())
	code, err := svc.Shorten(context.Background(), "https://go.dev")
	if err != nil{
		t.Fatalf("tidak boleh error, dapat: %v", err)
	}
	if len(code)!=7{
		t.Fatalf("panjang code harus 7, dapat: %v", len(code))
	}
}
func TestShorten_InvalidURL(t *testing.T){
	svc := New(newFakeRepo())

	_, err:= svc.Shorten(context.Background(), "javascript:alert(1)")
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