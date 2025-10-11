# Simple Rate Limiter

A production-ready HTTP rate limiter implementation in Go showcasing multiple rate limiting strategies through a clean Strategy pattern interface.

## Overview

This project implements a middleware-based rate limiting system for HTTP servers. It provides five different rate limiting algorithms, each with distinct characteristics and trade-offs. The implementation uses Go's native concurrency primitives and follows idiomatic Go patterns.

## Architecture

- **Strategy Interface**: Pluggable design allowing easy algorithm swapping
- **Middleware Pattern**: Standard HTTP middleware for transparent request interception
- **Per-IP Rate Limiting**: Tracks and limits requests based on client IP addresses
- **Thread-Safe**: All implementations handle concurrent access safely

## Pending

- **Concurrent Request Testing**: Load testing with actual concurrent requests to validate thread-safety and performance under pressure
- **Stats Improvements**: Fix inconsistent `Stats()` behavior across strategies and remove side effects
- **Memory Cleanup**: Implement TTL-based cleanup for inactive IPs to prevent memory leaks

## Implemented Strategies

### 1. Fixed Window

Divides time into fixed windows and counts requests within each window.

### 2. Sliding Window Log

Maintains exact timestamps of all requests within the window.

### 3. Sliding Window Counter

Hybrid approach using weighted counts from current and previous windows.

### 4. Token Bucket

Refills tokens at a constant rate; requests consume tokens.

### 5. Leaky Bucket

Processes requests from a bucket at a fixed rate using a background goroutine.

## Strategy Comparison

| Strategy                   | Memory Usage   | Accuracy               | Burst Handling            |
| -------------------------- | -------------- | ---------------------- | ------------------------- |
| **Fixed Window**           | O(n) - Low     | ⚠️ Burst at boundaries | Poor - Allows 2x burst    |
| **Sliding Window Log**     | O(n\*m) - High | ✅ Exact               | Good - Precise tracking   |
| **Sliding Window Counter** | O(n) - Low     | ✅ Approximate         | Good - Smooth             |
| **Token Bucket**           | O(n) - Low     | ✅ Exact               | Excellent - Natural burst |
| **Leaky Bucket**           | O(n) - Low     | ✅ Exact               | Poor - Strict rate        |

**Legend:**

- `n` = number of unique IPs
- `m` = limit per window (for Sliding Window Log)

## Detailed Strategy Analysis

### Fixed Window

**Pros:**

- Extremely simple to implement and understand
- Minimal memory overhead
- Fast lookups and updates
- Easy to reason about

**Cons:**

- Boundary burst problem: Users can make 2x requests at window boundaries
- Not smooth - sharp resets at window boundaries
- Less accurate rate limiting

---

### Sliding Window Log

**Pros:**

- Perfectly accurate - no boundary issues
- Smooth rate limiting across time
- Easy to debug with exact request logs

**Cons:**

- Memory-intensive (stores every timestamp)
- Requires cleanup of old timestamps on every request
- O(m) complexity per request for filtering

---

### Sliding Window Counter

**Pros:**

- Memory efficient like Fixed Window
- Smooth like Sliding Window Log
- Good approximation with low overhead
- Best of both worlds

**Cons:**

- Slightly less accurate (weighted estimate)
- More complex math than Fixed Window
- Can still allow minor bursts

---

### Token Bucket

**Pros:**

- Naturally handles burst traffic
- Smooth token refill
- Industry-standard algorithm
- Flexible - easy to tune burst vs sustained rate

**Cons:**

- Conceptually more complex
- Floating-point arithmetic (minor precision issues)
- Harder to predict exact behavior

---

### Leaky Bucket

**Pros:**

- Guarantees perfectly smooth output rate
- No bursts - strict rate enforcement
- Good for downstream protection

**Cons:**

- Strictest algorithm - rejects legitimate bursts
- Requires background goroutine (resource overhead)
- More complex cleanup and shutdown
- Poor user experience under variable load

## Key Learnings

### Concurrency Patterns

- **Mutex vs Channels**: Most strategies use `sync.Mutex` for simplicity; Leaky Bucket demonstrates channel-based coordination
- **Race Conditions**: Critical to identify shared state and protect all access paths

### Algorithm Trade-offs

- **No Perfect Solution**: Each algorithm optimizes for different constraints
- **Boundary Problems**: Understanding edge cases (window boundaries, float precision) is critical
- **Memory vs Accuracy**: Sliding Window Log vs Sliding Window Counter exemplifies this trade-off

### Design Patterns

- **Strategy Pattern**: Clean separation enables easy testing and swapping
- **Middleware Pattern**: Standard Go HTTP pattern for cross-cutting concerns
- **Constructor Pattern**: `New*` functions encapsulate initialization complexity

### Production Considerations

1. **Memory Leaks**: Need cleanup for inactive IPs (TTL-based eviction)
2. **Error Handling**: Don't ignore errors from standard library functions
3. **Observability**: Proper logging/metrics over debug prints
4. **Graceful Shutdown**: Context propagation for goroutine cleanup

### Real-World Insights

- **Fixed Window**: Simple but dangerous - doubles effective rate at boundaries
- **Token Bucket**: Most versatile - allows burst while protecting sustained rate
- **Sliding Window Counter**: Best general-purpose choice for most APIs
- **Leaky Bucket**: Overengineered for most use cases unless downstream has strict requirements

## Technical Highlights

- Zero external dependencies for rate limiting logic (except `godotenv` for config)
- Idiomatic Go patterns throughout
- Clean separation of concerns
- Easy to extend with new strategies
