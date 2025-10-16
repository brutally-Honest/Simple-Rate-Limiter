package ratelimiter

type Strategy interface {
	Allow(ip string) bool
	Close()
}
