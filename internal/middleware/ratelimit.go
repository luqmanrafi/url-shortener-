package middleware

import (
	"net"
	"net/http"
	// "strings"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	mu sync.Mutex
	visitors map[string]*rate.Limiter
	rps rate.Limit
	burst int
}

func NewRateLimiter(rps float64, burst int) *RateLimiter{
	return &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rps: rate.Limit(rps),
		burst: burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter{
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, ada := rl.visitors[ip]
	if !ada {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.visitors[ip] = limiter
	}
	return limiter
}

func (rl *RateLimiter) Middleware(next http.Handler)http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !rl.getLimiter(ip).Allow(){
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}