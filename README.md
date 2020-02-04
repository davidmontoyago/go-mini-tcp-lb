# go-tcp-termination-lb

Prototype edge/middle termination TCP load balancer with round robin balancing. For learning purposes only.

From [Introduction to modern network load balancing and proxying](https://blog.envoyproxy.io/introduction-to-modern-network-load-balancing-and-proxying-a57f6ff80236):

> Connection termination in close proximity (low latency) to the client has substantial performance implications. Specifically, if a terminating load balancer can be placed close to clients that are using a lossy network (e.g., cellular), retransmits are likely to happen faster prior to the data being moved to reliable fiber transit en-route to its ultimate location.

## Run me

```
# run upstream servers
while true; do { echo -e "hello back from 9001" | ncat -l 9001; test $? -gt 128 && break; } done & 
    while true; do { echo -e "hello back from 9002" | ncat -l 9002; test $? -gt 128 && break; } done &
        while true; do { echo -e "hello back from 9003" | ncat -l 9003; test $? -gt 128 && break; } done &


# run load balancer
go run -race *.go

# run concurrent clients
for i in {1..100}; do echo "hello $i" | nc localhost 9000; done & for i in {101..200}; do echo "hello $i" | nc localhost 9000; done

# clean up
pkill ncat -l 900
```