package cmdrunr

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

func WaitForPort(ctx context.Context, port int) bool {
	if port < 0 || port > 65535 {
		// Invalid port is considered ok
		log.Printf("Warning: an invalid port was specified")
		return true
	}

	dialContext, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for !checkPort(dialContext, port) {
		select {
		case <-dialContext.Done():
			return false
		case <-time.After(500 * time.Millisecond):
		}
	}

	return true
}

func checkPort(ctx context.Context, port int) bool {
	d := net.Dialer{Timeout: 60 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)))

	if err == nil {
		defer conn.Close()
		return true
	}

	return false
}
