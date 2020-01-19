# go-mini-tcp-lb

Prototype miniature `termination TCP load balancer` with round robin balancing. For learning purposes only.

From [Introduction to modern network load balancing and proxying](https://blog.envoyproxy.io/introduction-to-modern-network-load-balancing-and-proxying-a57f6ff80236):

> Connection termination in close proximity (low latency) to the client has substantial performance implications. Specifically, if a terminating load balancer can be placed close to clients that are using a lossy network (e.g., cellular), retransmits are likely to happen faster prior to the data being moved to reliable fiber transit en-route to its ultimate location.

## Run me

```
# run upstream servers
nc -lk 9001 & nc -lk 9002 & nc -lk 9003

# run load balancer
# go run -race *.go

# run concurrent clients
for i in {1..100}; do echo "hello $i" | nc localhost 9000; done & for i in {101..200}; do echo "hello $i" | nc localhost 9000; done

# clean up
pkill nc -lk
```