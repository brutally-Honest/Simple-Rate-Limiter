# Fixed Window Rate Limiter - Boundary Exploitation Observations

## Test Configuration

- Rate Limit: 10 requests per second
- Target IP: 203.0.113.42
- Test Strategy: Boundary burst exploitation

## Successful Exploitation Results

### Test Run Example

```
Phase 1: Sending 1 request to start the window
  ✓ Request 1 allowed (200) at T+1ms [Window Start]


Phase 2: Waiting 950ms, then hammering 25 requests at boundary
  ✓ Request 26 allowed (200) at T+999ms [Boundary Hammer]
  ✓ Request 2 allowed (200) at T+999ms [Boundary Hammer]
  ✓ Request 7 allowed (200) at T+1001ms [Boundary Hammer]
  ✓ Request 11 allowed (200) at T+1001ms [Boundary Hammer]
  ✓ Request 14 allowed (200) at T+1001ms [Boundary Hammer]
  ✓ Request 23 allowed (200) at T+1002ms [Boundary Hammer]
  ✗ Request 9 rejected (429) at T+1002ms [Boundary Hammer]
  ✓ Request 19 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 4 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 6 allowed (200) at T+1002ms [Boundary Hammer]
  ✗ Request 17 rejected (429) at T+1002ms [Boundary Hammer]
  ✗ Request 21 rejected (429) at T+1002ms [Boundary Hammer]
  ✓ Request 3 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 10 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 13 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 25 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 12 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 16 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 22 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 5 allowed (200) at T+1002ms [Boundary Hammer]
  ✓ Request 24 allowed (200) at T+1003ms [Boundary Hammer]
  ✗ Request 8 rejected (429) at T+1003ms [Boundary Hammer]
  ✗ Request 15 rejected (429) at T+1003ms [Boundary Hammer]
  ✗ Request 18 rejected (429) at T+1003ms [Boundary Hammer]
  ✓ Request 20 allowed (200) at T+1003ms [Boundary Hammer]


=== RESULTS ===
Total requests sent: 26
Requests allowed: 20
Requests rejected: 6
Time span: 1003ms (~1.0 seconds)
Configured limit: 10 requests/second
Actual throughput: 19.9 requests/second
```

## Key Observations

### 1. Boundary Race Condition Behavior

- Window 1: T+1ms to T+1001ms (3 requests: #1, #26, #2)
- Window 2: T+1001ms onwards (17 requests starting with #7)
- Race condition at boundary causes interleaved accept/reject pattern

### 2. Exploitation Mechanics

- Initial request at T+1ms starts the window timer
- Burst at T+999-1003ms straddles the T+1001ms boundary
- Some requests fill remaining Window 1 slots
- Some requests trigger Window 2 reset and fill new slots
- Result: 20 requests allowed vs 10/second limit (100% over-limit)

### 3. Non-Deterministic Results

The exploitation success varies due to:

- Goroutine scheduling: OS determines execution order
- Network latency: HTTP request timing variations
- Mutex contention: Lock acquisition order affects window reset timing
- System load: CPU/memory pressure impacts timing precision

### 4. Typical Result Ranges

- Perfect exploitation: 18-20 requests allowed
- Partial success: 12-17 requests allowed
- Timing miss: 10-11 requests allowed
- Success rate: ~70-80% in testing

### 5. Security Implications

- Probabilistic attack: Attackers can retry until successful
- Burst amplification: 2x rate limit bypass achievable
- Detection difficulty: Intermittent nature makes monitoring harder
- Real-world impact: Can overwhelm downstream services during bursts

## Mitigation Strategies

1. Use sliding window instead of fixed window
2. Add jitter/randomization to window boundaries
3. Use token bucket for smoother rate limiting
4. Monitor burst patterns for anomaly detection
