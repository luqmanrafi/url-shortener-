package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log"
	"math/big"
	"net/url"

	"github.com/luqmanrafi/urlshortener/internal/model"
)

const (
	alphabet   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	codeLength = 7
	maxRetries = 5
)

type Repository interface {
	Save(ctx context.Context, code, originalURL string) error
	FindByCode(ctx context.Context, code string)(string, error)
	IncrementClick(ctx context.Context, code string)error
	GetStats(ctx context.Context, code string)(*model.URL, error)
	Ping(ctx context.Context)error
}

type Shortener struct{
	repo Repository
}

func New(repo Repository)*Shortener{
	return &Shortener{repo: repo}
}

func (s *Shortener) Shorten(ctx context.Context, rawURL string)(string, error){
	if err:= validateURL(rawURL); err != nil{
		return "", err
	}

	for i := 0; i < maxRetries; i++{
		code, err := generateCode()
		if err != nil{
			return "", err
		}
		err = s.repo.Save(ctx, code, rawURL)
		if err == nil{
			return code, nil
		}
		if errors.Is(err, model.ErrCodeExists){
			continue
		}
		return "", err
	}
	return "", errors.New("Failed to generate unique code")
}

func (s *Shortener) Resolve(ctx context.Context, code string)(string, error){
	original, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		return "", err
	}
	if err := s.repo.IncrementClick(ctx, code); err != nil{
		log.Printf("increment click %s: %v", code , err)
	}
	return original, err
}

func (s *Shortener) Stats(ctx context.Context, code string)(*model.URL, error){
	return s.repo.GetStats(ctx, code)
}

func (s *Shortener) Ready(ctx context.Context)error{
	return s.repo.Ping(ctx)
}

func validateURL(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return model.ErrInvalidUrl
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return model.ErrInvalidUrl
	}
	return nil
}

func generateCode() (string, error){
	b := make([]byte, codeLength)
	max := big.NewInt(int64(len(alphabet)))
	for i := range b{
		n, err := rand.Int(rand.Reader, max)
		if err != nil{
			return "", err
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}