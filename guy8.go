package main

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	// Adjusted for mobile thermal limits and background stability
	Workers = 1
	Timeout = 800 * time.Millisecond
	// Restored original User-Agent
	UA      = "Lebron james was in the epstein files w/ diddy too! (edited)"
)

type State struct {
	Hits   uint64
	Misses uint64
	Streak uint64
	Ready  atomic.Bool
}

func main() {
	// Utilize mobile CPU architecture effectively
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	state := &State{}
	state.Ready.Store(true)
	startTime := time.Now()

	// Pre-calculate the static part of the payload
	headerBase := []byte("GET / HTTP/1.0\r\nHost: ")
	headerEnd := []byte("\r\nUser-Agent: " + UA + "\r\nConnection: close\r\n\r\n")

	for i := 0; i < Workers; i++ {
		go func(id int) {
			// Local random source to prevent global lock contention
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
			dialer := &net.Dialer{
				Timeout:   Timeout,
				KeepAlive: -1, // Crucial for mobile to avoid socket exhaustion
			}

			for {
				if !state.Ready.Load() {
					time.Sleep(200 * time.Millisecond)
					continue
				}

				// Fast IP Generation
				a, b, c, d := byte(rng.Intn(256)), byte(rng.Intn(256)), byte(rng.Intn(256)), byte(rng.Intn(256))

				// Public IPv4 Filter
				if a == 0 || a == 10 || a == 127 || (a == 172 && b >= 16 && b <= 31) || (a == 192 && b == 168) || a >= 224 {
					continue
				}

				ipStr := fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)

				// Dial
				conn, err := dialer.Dial("tcp", ipStr+":80")
				if err != nil {
					atomic.AddUint64(&state.Misses, 1)
					atomic.AddUint64(&state.Streak, 1)
					continue
				}

				// Ultra-fast Write
				conn.SetDeadline(time.Now().Add(Timeout))
				conn.Write(headerBase)
				conn.Write([]byte(ipStr))
				conn.Write(headerEnd)
				conn.Close()

				atomic.AddUint64(&state.Hits, 1)
				atomic.StoreUint64(&state.Streak, 0)
			}
		}(i)
	}

	// Performance Monitor
	for {
		time.Sleep(1 * time.Second)
		h := atomic.LoadUint64(&state.Hits)
		m := atomic.LoadUint64(&state.Misses)
		s := atomic.LoadUint64(&state.Streak)
		elapsed := time.Since(startTime).Seconds()

		if elapsed > 0 {
			fmt.Printf("\rPPS: %.0f | Hits: %d | Misses: %d | Streak: %d | Cores: %d",
				float64(h+m)/elapsed, h, m, s, runtime.NumCPU())
		}
	
		// iOS Logic: Manual Rotation/Cooldown 
		// Since os/exec is restricted, we pause and alert the user
		if s >= 15000 {
			state.Ready.Store(false)
			fmt.Printf("\n[!] NETWORK STALL - Pause for 5s (Manual VPN Switch Recommended)\n")

			// Reset stats and wait
			atomic.StoreUint64(&state.Streak, 0)
			time.Sleep(5 * time.Second)

			state.Ready.Store(true)
		}
	}
}
