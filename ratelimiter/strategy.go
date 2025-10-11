package ratelimiter

type Strategy interface {
	Allow(ip string) bool
	Stats(ip string) int
}
