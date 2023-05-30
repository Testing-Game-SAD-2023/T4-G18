package limiter

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ClientLimiter struct {
	mu         sync.RWMutex
	clients    map[string]*client
	burst      int
	bucketSize int
}

func NewClientLimiter(burst, bucketSize int) *ClientLimiter {
	return &ClientLimiter{
		mu:         sync.RWMutex{},
		clients:    make(map[string]*client),
		burst:      burst,
		bucketSize: bucketSize,
	}
}

func (cm *ClientLimiter) Cleanup(timeout time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for ip, client := range cm.clients {
		if time.Since(client.lastSeen) > timeout {
			delete(cm.clients, ip)
		}
	}
}

func (cm *ClientLimiter) get(ip string) (*rate.Limiter, bool) {

	cm.mu.RLock()
	defer cm.mu.RUnlock()
	v := cm.clients[ip]
	if v != nil {
		v.lastSeen = time.Now()
		return v.limiter, true
	}
	return nil, false
}

func (cm *ClientLimiter) add(ip string) *rate.Limiter {

	cm.mu.Lock()
	defer cm.mu.Unlock()
	limiter := rate.NewLimiter(rate.Limit(cm.burst), cm.bucketSize)
	cm.clients[ip] = &client{
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	return limiter
}

func (cm *ClientLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the IP address for the current user.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Call the getVisitor function to retreive the rate limiter for
		// the current user.
		limiter, ok := cm.get(ip)
		if !ok {
			limiter = cm.add(ip)
		}
		if !limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
