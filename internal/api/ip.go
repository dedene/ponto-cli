package api

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// detectOutboundIP tries to detect the client's outbound IP address.
// This is required for PSD2 compliance when creating synchronizations.
func detectOutboundIP(ctx context.Context) string {
	// Try external service first
	if ip := detectIPFromService(ctx); ip != "" {
		return ip
	}

	// Fall back to local interface detection
	return detectLocalIP()
}

func detectIPFromService(ctx context.Context) string {
	client := &http.Client{Timeout: 5 * time.Second}

	// Try multiple services for redundancy
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
	}

	for _, svc := range services {
		reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, svc, nil)
		if err != nil {
			cancel()

			continue
		}

		resp, err := client.Do(req)
		cancel()
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err != nil {
			continue
		}

		ip := strings.TrimSpace(string(body))
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	return ""
}

func detectLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "0.0.0.0"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
