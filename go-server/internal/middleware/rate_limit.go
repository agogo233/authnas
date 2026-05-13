package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/gin-gonic/gin"
)

type visitor struct {
	windowStart  time.Time
	lastSeen     time.Time
	requestCount int
}

var (
	visitors           = make(map[string]*visitor)
	mu                 sync.RWMutex
	cleanupInterval    = time.Minute
	stopCleanupChannel chan struct{}
	cleanupDoneChannel chan struct{}
	maxVisitors        = 100000
)

func RateLimit(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.RateLimit.Enabled {
			c.Next()
			return
		}

		ip := GetClientIP(c)
		requestsPerMinute := cfg.RateLimit.RequestsPerMinute
		if requestsPerMinute <= 0 {
			requestsPerMinute = 60
		}

		mu.Lock()
		defer mu.Unlock()

		if len(visitors) >= maxVisitors {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit system overloaded"})
			c.Abort()
			return
		}

		now := time.Now()
		windowStart := now.Truncate(time.Minute)
		v, exists := visitors[ip]

		if !exists || v.windowStart.Before(windowStart) {
			visitors[ip] = &visitor{
				windowStart:  windowStart,
				lastSeen:     now,
				requestCount: 1,
			}
			c.Next()
			return
		}

		v.requestCount++
		v.lastSeen = now

		if v.requestCount > requestsPerMinute {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func cleanupVisitors(stop <-chan struct{}, done chan<- struct{}) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > cleanupInterval*5 {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		case <-stop:
			close(done)
			return
		}
	}
}

func RateLimitByUser(cfg *config.Config, userID string) bool {
	if !cfg.RateLimit.Enabled {
		return true
	}

	key := "user:" + userID
	requestsPerMinute := cfg.RateLimit.RequestsPerMinute
	if requestsPerMinute <= 0 {
		requestsPerMinute = 60
	}
	windowDuration := time.Minute

	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	v, exists := visitors[key]

	if !exists {
		visitors[key] = &visitor{
			windowStart:  now,
			lastSeen:     now,
			requestCount: 1,
		}
		return true
	}

	if now.Sub(v.windowStart) >= windowDuration {
		v.windowStart = now
		v.requestCount = 1
		v.lastSeen = now
		return true
	}

	v.requestCount++
	v.lastSeen = now

	if v.requestCount > requestsPerMinute {
		return false
	}

	return true
}

var (
	passwordResetVisitors      = make(map[string]*passwordResetAttempt)
	passwordResetMu            sync.RWMutex
	passwordResetLimitPerEmail = 3
	passwordResetLimitPerIP    = 10
	passwordResetWindow        = 15 * time.Minute
	passwordResetStopChan      = make(chan struct{})
	passwordResetDoneChan      = make(chan struct{})
)

var (
	resetCodeVisitors     = make(map[string]*resetCodeAttempt)
	resetCodeMu           sync.RWMutex
	resetCodeLimitPerIP   = 10
	resetCodeLimitPerCode = 5
	resetCodeWindow       = 15 * time.Minute
	resetCodeStopChan     = make(chan struct{})
	resetCodeDoneChan     = make(chan struct{})
)

type resetCodeAttempt struct {
	codeAttempts map[string]int
	ipAttempts   int
	firstAttempt time.Time
}

var (
	inviteVisitors      = make(map[string]*inviteAttempt)
	inviteMu            sync.RWMutex
	inviteLimitPerIP    = 10
	inviteLimitPerEmail = 5
	inviteWindow        = 15 * time.Minute
)

type inviteAttempt struct {
	emailAttempts int
	ipAttempts    int
	firstAttempt  time.Time
}

var (
	emailVerifyVisitors      = make(map[string]*emailVerifyAttempt)
	emailVerifyMu            sync.RWMutex
	emailVerifyLimitPerIP    = 10
	emailVerifyLimitPerEmail = 3
	emailVerifyWindow        = 15 * time.Minute
	emailVerifyStopChan      = make(chan struct{})
	emailVerifyDoneChan      = make(chan struct{})
)

type emailVerifyAttempt struct {
	emailAttempts int
	ipAttempts    int
	firstAttempt  time.Time
}

type passwordResetAttempt struct {
	emailAttempts    int
	ipAttempts       int
	firstAttemptTime time.Time
}

func ResetPasswordResetRateLimit() {
	passwordResetMu.Lock()
	defer passwordResetMu.Unlock()
	passwordResetVisitors = make(map[string]*passwordResetAttempt)
}

func CheckPasswordResetRateLimit(email, ip string) (bool, string) {
	passwordResetMu.Lock()
	defer passwordResetMu.Unlock()

	now := time.Now()
	key := ip

	v, exists := passwordResetVisitors[key]
	if !exists {
		v = &passwordResetAttempt{
			firstAttemptTime: now,
		}
		passwordResetVisitors[key] = v
	}

	if now.Sub(v.firstAttemptTime) > passwordResetWindow {
		v.emailAttempts = 0
		v.ipAttempts = 0
		v.firstAttemptTime = now
	}

	v.ipAttempts++
	if v.ipAttempts > passwordResetLimitPerIP {
		return false, "too many password reset attempts from this IP"
	}

	if email != "" && v.emailAttempts >= passwordResetLimitPerEmail && now.Sub(v.firstAttemptTime) <= passwordResetWindow {
		return false, "too many password reset attempts for this email"
	}

	return true, ""
}

func RecordPasswordResetAttempt(email, ip string) {
	if email == "" {
		return
	}

	passwordResetMu.Lock()
	defer passwordResetMu.Unlock()

	now := time.Now()
	key := ip

	v, exists := passwordResetVisitors[key]
	if !exists {
		v = &passwordResetAttempt{
			firstAttemptTime: now,
		}
		passwordResetVisitors[key] = v
	}

	if now.Sub(v.firstAttemptTime) > passwordResetWindow {
		v.emailAttempts = 0
		v.ipAttempts = 0
		v.firstAttemptTime = now
	}

	v.emailAttempts++
}

func CheckInviteRateLimit(email, ip string) (bool, string) {
	inviteMu.Lock()
	defer inviteMu.Unlock()

	now := time.Now()
	v, exists := inviteVisitors[ip]

	if !exists {
		inviteVisitors[ip] = &inviteAttempt{firstAttempt: now}
	}

	if now.Sub(v.firstAttempt) > inviteWindow {
		v.ipAttempts = 0
		v.emailAttempts = 0
		v.firstAttempt = now
	}

	v.ipAttempts++
	if v.ipAttempts > inviteLimitPerIP {
		return false, "too many registration attempts from this IP"
	}

	if email != "" {
		v.emailAttempts++
		if v.emailAttempts > inviteLimitPerEmail {
			return false, "too many registration attempts for this email"
		}
	}

	return true, ""
}

func RecordInviteAttempt(email, ip string) {
	inviteMu.Lock()
	defer inviteMu.Unlock()

	now := time.Now()
	v, exists := inviteVisitors[ip]

	if !exists {
		v = &inviteAttempt{firstAttempt: now}
		inviteVisitors[ip] = v
	}

	if now.Sub(v.firstAttempt) > inviteWindow {
		v.ipAttempts = 0
		v.emailAttempts = 0
		v.firstAttempt = now
	}
}

func CheckResetCodeRateLimit(code, ip string) (bool, string) {
	resetCodeMu.Lock()
	defer resetCodeMu.Unlock()

	now := time.Now()
	v, exists := resetCodeVisitors[ip]
	if !exists {
		v = &resetCodeAttempt{
			codeAttempts: make(map[string]int),
			firstAttempt: now,
		}
		resetCodeVisitors[ip] = v
	}

	if now.Sub(v.firstAttempt) > resetCodeWindow {
		v.codeAttempts = make(map[string]int)
		v.ipAttempts = 0
		v.firstAttempt = now
	}

	v.ipAttempts++
	if v.ipAttempts > resetCodeLimitPerIP {
		return false, "too many reset attempts from this IP"
	}

	v.codeAttempts[code]++
	if v.codeAttempts[code] > resetCodeLimitPerCode {
		return false, "too many attempts for this code"
	}

	return true, ""
}

func CheckEmailVerifyRateLimit(email, ip string) (bool, string) {
	emailVerifyMu.Lock()
	defer emailVerifyMu.Unlock()

	now := time.Now()
	v, exists := emailVerifyVisitors[ip]
	if !exists {
		v = &emailVerifyAttempt{
			firstAttempt: now,
		}
		emailVerifyVisitors[ip] = v
	}

	if now.Sub(v.firstAttempt) > emailVerifyWindow {
		v.emailAttempts = 0
		v.ipAttempts = 0
		v.firstAttempt = now
	}

	v.ipAttempts++
	if v.ipAttempts > emailVerifyLimitPerIP {
		return false, "too many email verification attempts from this IP"
	}

	if email != "" {
		v.emailAttempts++
		if v.emailAttempts > emailVerifyLimitPerEmail {
			return false, "too many email verification attempts for this email"
		}
	}

	return true, ""
}

func RecordEmailVerifyAttempt(email, ip string) {
	emailVerifyMu.Lock()
	defer emailVerifyMu.Unlock()

	now := time.Now()
	v, exists := emailVerifyVisitors[ip]
	if !exists {
		v = &emailVerifyAttempt{
			firstAttempt: now,
		}
		emailVerifyVisitors[ip] = v
	}

	if now.Sub(v.firstAttempt) > emailVerifyWindow {
		v.emailAttempts = 0
		v.ipAttempts = 0
		v.firstAttempt = now
	}
}

func ResetEmailVerifyRateLimit() {
	emailVerifyMu.Lock()
	defer emailVerifyMu.Unlock()
	emailVerifyVisitors = make(map[string]*emailVerifyAttempt)
}

func RecordResetCodeAttempt(code, ip string) {
	if code == "" {
		return
	}

	resetCodeMu.Lock()
	defer resetCodeMu.Unlock()

	now := time.Now()
	v, exists := resetCodeVisitors[ip]
	if !exists {
		v = &resetCodeAttempt{
			codeAttempts: make(map[string]int),
			firstAttempt: now,
		}
		resetCodeVisitors[ip] = v
	}

	if now.Sub(v.firstAttempt) > resetCodeWindow {
		v.codeAttempts = make(map[string]int)
		v.ipAttempts = 0
		v.firstAttempt = now
	}

	v.codeAttempts[code]++
}

func ResetResetCodeRateLimit(code, ip string) {
	resetCodeMu.Lock()
	defer resetCodeMu.Unlock()

	delete(resetCodeVisitors[ip].codeAttempts, code)
}

func init() {
	stopCleanupChannel = make(chan struct{})
	cleanupDoneChannel = make(chan struct{})
	go cleanupVisitors(stopCleanupChannel, cleanupDoneChannel)
	go cleanupPasswordResetVisitors(passwordResetStopChan, passwordResetDoneChan)
	go cleanupResetCodeVisitors(resetCodeStopChan, resetCodeDoneChan)
	go cleanupEmailVerifyVisitors(emailVerifyStopChan, emailVerifyDoneChan)
}

func cleanupPasswordResetVisitors(stop <-chan struct{}, done chan<- struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			passwordResetMu.Lock()
			now := time.Now()
			cutoff := now.Add(-passwordResetWindow * 2)
			for ip, v := range passwordResetVisitors {
				if v.firstAttemptTime.Before(cutoff) {
					delete(passwordResetVisitors, ip)
				}
			}
			passwordResetMu.Unlock()
		case <-stop:
			close(done)
			return
		}
	}
}

func cleanupResetCodeVisitors(stop <-chan struct{}, done chan<- struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			resetCodeMu.Lock()
			now := time.Now()
			cutoff := now.Add(-resetCodeWindow * 2)
			for ip, v := range resetCodeVisitors {
				if v.firstAttempt.Before(cutoff) {
					delete(resetCodeVisitors, ip)
				}
			}
			resetCodeMu.Unlock()
		case <-stop:
			close(done)
			return
		}
	}
}

func cleanupEmailVerifyVisitors(stop <-chan struct{}, done chan<- struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			emailVerifyMu.Lock()
			now := time.Now()
			cutoff := now.Add(-emailVerifyWindow * 2)
			for ip, v := range emailVerifyVisitors {
				if v.firstAttempt.Before(cutoff) {
					delete(emailVerifyVisitors, ip)
				}
			}
			emailVerifyMu.Unlock()
		case <-stop:
			close(done)
			return
		}
	}
}
