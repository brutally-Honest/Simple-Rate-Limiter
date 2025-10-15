# Rate Limiting Strategies

## Fixed Window

| Aspect   | Details                                                                                                                                                        |
| -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Pros     | - Extremely simple to implement<br>- Minimal memory overhead<br>- Fast lookups and updates                                                                     |
| Cons     | - Boundary burst problem: Users can make 2x requests at window boundaries<br>- Not smooth - sharp resets at window boundaries<br>- Less accurate rate limiting |
| Use When | Simplicity trumps accuracy, and traffic is relatively uniform                                                                                                  |

---

## Sliding Window Log

| Aspect   | Details                                                                                                                                             |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Pros     | - Perfectly accurate - no boundary issues<br>- Smooth rate limiting across time<br>- Easy to debug with exact request logs                          |
| Cons     | - Memory-intensive (stores every timestamp)<br>- Requires cleanup of old timestamps on every request<br>- O(m) complexity per request for filtering |
| Use When | Accuracy is critical and memory isn't a constraint                                                                                                  |

---

## Sliding Window Counter

| Aspect   | Details                                                                                                                                     |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| Pros     | - Memory efficient like Fixed Window<br>- Smooth like Sliding Window Log<br>- Good approximation with low overhead<br>- Best of both worlds |
| Cons     | - Slightly less accurate (weighted estimate)<br>- More complex math than Fixed Window<br>- Can still allow minor bursts                     |
| Use When | General-purpose API rate limiting (best default choice)                                                                                     |

---

## Token Bucket

| Aspect   | Details                                                                                                                                          |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| Pros     | - Naturally handles burst traffic<br>- Smooth token refill<br>- Industry-standard algorithm<br>- Flexible - easy to tune burst vs sustained rate |
| Cons     | - Conceptually more complex<br>- Floating-point arithmetic (minor precision issues)<br>- Harder to predict exact behavior                        |
| Use When | Burst tolerance is important, API needs to feel responsive                                                                                       |

---

## Leaky Bucket

| Aspect   | Details                                                                                                                                                                                       |
| -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Pros     | - Guarantees perfectly smooth output rate<br>- No bursts - strict rate enforcement<br>- Good for downstream protection                                                                        |
| Cons     | - Strictest algorithm - rejects legitimate bursts<br>- Requires background goroutine (resource overhead)<br>- More complex cleanup and shutdown<br>- Poor user experience under variable load |
| Use When | Protecting fragile downstream systems requiring strict rate control                                                                                                                           |

---

## Quick Comparison

| Strategy               | Memory  | Accuracy | Burst Handling            | Complexity |
| ---------------------- | ------- | -------- | ------------------------- | ---------- |
| Fixed Window           | O(n)    | ⚠️ Low   | Poor - Allows 2x burst    | Simple     |
| Sliding Window Log     | O(n\*m) | ✅ Exact | Good - Precise tracking   | Medium     |
| Sliding Window Counter | O(n)    | ✅ Good  | Good - Smooth             | Medium     |
| Token Bucket           | O(n)    | ✅ Exact | Excellent - Natural burst | Complex    |
| Leaky Bucket           | O(n)    | ✅ Exact | Poor - Strict rate        | Complex    |

Legend: n = unique IPs, m = limit per window

---

## Observations

Most APIs: Sliding Window Counter - great balance of accuracy, performance, and memory

Burst-tolerant: Token Bucket - natural burst handling while protecting sustained rate

Simple/High-throughput: Fixed Window - fast but watch for boundary abuse

Strict downstream protection: Leaky Bucket - smooth, predictable output rate
