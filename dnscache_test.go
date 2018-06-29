package dnscache

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestResolver_LookupHost(t *testing.T) {
	r := &Resolver{}
	var cacheMiss bool
	r.onCacheMiss = func() {
		cacheMiss = true
	}
	hosts := []string{"google.com", "google.com.", "netflix.com"}
	for _, host := range hosts {
		t.Run(host, func(t *testing.T) {
			for _, wantMiss := range []bool{true, false, false} {
				cacheMiss = false
				addrs, err := r.LookupHost(context.Background(), host)
				if err != nil {
					t.Fatal(err)
				}
				if len(addrs) == 0 {
					t.Error("got no record")
				}
				for _, addr := range addrs {
					if net.ParseIP(addr) == nil {
						t.Errorf("got %q; want a literal IP address", addr)
					}
				}
				if wantMiss != cacheMiss {
					t.Errorf("got cache miss=%v, want %v", cacheMiss, wantMiss)
				}
			}
		})
	}
}

func TestClearCache(t *testing.T) {
	r := &Resolver{}
	_, _ = r.LookupHost(context.Background(), "google.com")
	if e, _ := r.cache.Load("hgoogle.com"); e != nil && !e.(*cacheEntry).used {
		t.Error("cache entry used flag is false, want true")
	}
	r.Refresh(true)
	if e, _ := r.cache.Load("hgoogle.com"); e != nil && e.(*cacheEntry).used {
		t.Error("cache entry used flag is true, want false")
	}
	r.Refresh(true)
	if e, _ := r.cache.Load("hgoogle.com"); e != nil {
		t.Error("cache entry is not cleared")
	}
}

func Example() {
	r := &Resolver{}
	t := &http.Transport{
		DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
			separator := strings.LastIndex(addr, ":")
			ips, err := r.LookupHost(ctx, addr[:separator])
			if err != nil {
				return nil, err
			}
			for _, ip := range ips {
				conn, err = net.Dial(network, ip+addr[separator:])
				if err == nil {
					break
				}
			}
			return
		},
	}
	c := &http.Client{Transport: t}
	res, err := c.Get("http://httpbin.org/status/418")
	if err == nil {
		fmt.Println(res.StatusCode)
	}
	// Output: 418
}
