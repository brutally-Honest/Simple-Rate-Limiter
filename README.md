# Simple Rate Limiter

Production-ready HTTP rate limiter in Go with five different strategies using a clean Strategy pattern.

## Features

- Strategy Interface: Pluggable design for easy algorithm swapping
- Middleware Pattern: Standard HTTP middleware for transparent request interception
- Per-IP Rate Limiting: Tracks and limits requests based on client IP
- Thread-Safe: All implementations handle concurrent access safely
- Memory Leak Prevention: Automatic cleanup of inactive IPs to prevent memory accumulation

## Strategies

| Strategy               | Description                                                    |
| ---------------------- | -------------------------------------------------------------- |
| Fixed Window           | Divides time into fixed windows and counts requests per window |
| Sliding Window Log     | Maintains exact timestamps of all requests within the window   |
| Sliding Window Counter | Hybrid using weighted counts from current and previous windows |
| Token Bucket           | Refills tokens at a constant rate; requests consume tokens     |
| Leaky Bucket           | Processes requests from a bucket at a fixed rate (background)  |

For detailed pros/cons and use cases, see [Rate Limiting Strategies](ratelimiter/strategy.md)

## Key Learnings

### Concurrency

- Mutex vs Channels: Most strategies use sync.Mutex for simplicity; Leaky Bucket demonstrates channel-based coordination
- Context Management: context.WithCancel() for graceful goroutine shutdown across all strategies
- Race Conditions: Critical to identify shared state and protect all access paths

### Design Patterns

- Strategy Pattern: Clean separation enables easy testing and swapping
- Middleware Pattern: Standard Go HTTP pattern for cross-cutting concerns
- Constructor Pattern: New\* functions encapsulate initialization complexity

### Rate Limiting Algorithm Tradeoffs

- No Perfect Solution: Each algorithm optimizes for different constraints
- Boundary Problems: Window boundaries and float precision edge cases matter
- Memory vs Accuracy: Sliding Window Log vs Counter exemplifies this trade-off

## Future Enhancements

- Unit Tests: Add comprehensive tests for all rate limiting strategies, including edge cases, concurrent access patterns, and boundary conditions
- Configurable memory cleanup: Add configurable cleanup intervals and max idle times for better memory management across all strategies
- Distributed Rate Limiting: Current implementation stores rate limit data locally per server instance, consider adding Redis/database coordination for shared state across multiple servers

## Technical Notes

- Zero external dependencies for rate limiting logic
- Clean separation of concerns
- Idiomatic Go patterns
- Easy to extend with new strategies
