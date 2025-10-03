# Balancer

[![Go Report Card](https://goreportcard.com/badge/github.com/NovikovRoman/balancer)](https://goreportcard.com/report/github.com/NovikovRoman/balancer)

A library for balancing the load on services.

## Usage

```go
package main

import (
    "context"
    "net/http"
    "time"

    "github.com/NovikovRoman/balancer"
)

// Arbitrary structure.
type myService struct {
    client *http.Client
}

func main() {
    items := []*balancer.Item[myService]{}
    s1 := &myService{
        client: &http.Client{
            Transport: &http.Transport{
                // …
            },
        },
    }
    s2 := &myService{
        client: &http.Client{
            Transport: &http.Transport{
                // …
            },
        },
    }
    s3 := &myService{
        client: &http.Client{
            Transport: &http.Transport{
                // …
            },
        },
    }

    items = append(items, balancer.NewItem(s1, 10)) // No more than 10 requests per second.
    items = append(items, balancer.NewItem(s2, 20)) // No more than 20 requests per second.
    items = append(items, balancer.NewItem(s3, 30)) // No more than 30 requests per second.

    b := balancer.New(items)

    s := b.Acquire()
    if s != nil { // If there is a service available, then make a request.
        _, _ = s.client.Get("https://api.site.domain/path")
    }

    ctx := context.Background()
    // With expectation. 5 attempts, half a second waiting time between attempts.
    s = b.AcquireWait(ctx, 5, time.Second/2)
    if s != nil { // If there is a service available, then make a request.
        _, _ = s.client.Get("https://api.site.domain/path")
    }

    b.SetShuffle(true) // Items will be balanced in random order.
}
```
