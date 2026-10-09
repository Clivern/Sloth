// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package probe

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	// Passing means the target answered as expected.
	Passing = "passing"
	// Critical means the target failed or could not be reached.
	Critical = "critical"
	// MaxOutput caps the output returned for a probe.
	MaxOutput = 4096
)

// HTTPTarget is a full URL to probe with an HTTP GET.
type HTTPTarget struct {
	URL     string
	Timeout time.Duration
}

// TCPTarget is an instance address to probe with a TCP dial.
type TCPTarget struct {
	Address string
	Port    int
	Timeout time.Duration
}

// Result is the outcome of one probe.
type Result struct {
	Status string
	Output string
}

// HTTP sends a GET to the target and reports passing for a 2xx response.
func HTTP(ctx context.Context, target HTTPTarget) Result {
	if target.URL == "" {
		return Result{Status: Critical, Output: "invalid target: url is required"}
	}

	client := &http.Client{Timeout: target.Timeout}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		target.URL,
		nil,
	)
	if err != nil {
		output := err.Error()
		if len(output) > MaxOutput {
			output = output[:MaxOutput]
		}
		return Result{Status: Critical, Output: output}
	}

	resp, err := client.Do(req)
	if err != nil {
		output := err.Error()
		if len(output) > MaxOutput {
			output = output[:MaxOutput]
		}
		return Result{Status: Critical, Output: output}
	}
	defer resp.Body.Close()

	output := resp.Status
	if len(output) > MaxOutput {
		output = output[:MaxOutput]
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Result{Status: Critical, Output: output}
	}

	return Result{Status: Passing, Output: output}
}

// TCP dials the target and reports passing when the connection opens.
func TCP(ctx context.Context, target TCPTarget) Result {
	if target.Address == "" || target.Port == 0 {
		return Result{Status: Critical, Output: "invalid target: address and port are required"}
	}

	dialer := net.Dialer{Timeout: target.Timeout}
	conn, err := dialer.DialContext(
		ctx,
		"tcp",
		net.JoinHostPort(target.Address, strconv.Itoa(target.Port)),
	)
	if err != nil {
		output := err.Error()
		if len(output) > MaxOutput {
			output = output[:MaxOutput]
		}
		return Result{Status: Critical, Output: output}
	}
	defer conn.Close()

	return Result{Status: Passing, Output: "tcp connection established"}
}
