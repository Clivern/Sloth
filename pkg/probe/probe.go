// Copyright 2026 Cognit. All rights reserved.
// License can be found in the LICENSE file.

package probe

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/clivern/cognit/pkg/util"
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
	URL     string `validate:"required,url" label:"URL"`
	Timeout time.Duration
}

// TCPTarget is an instance address to probe with a TCP dial.
type TCPTarget struct {
	Address string `validate:"required,max=255" label:"Address"`
	Port    int    `validate:"required,min=1,max=65535" label:"Port"`
	Timeout time.Duration
}

// Result is the outcome of one probe.
type Result struct {
	Status string
	Output string
}

// Validate checks the target fields and returns the first validation error.
func (t HTTPTarget) Validate() error {
	return util.ValidateStruct(t)
}

// Validate checks the target fields and returns the first validation error.
func (t TCPTarget) Validate() error {
	return util.ValidateStruct(t)
}

// HTTP sends a GET to the target and reports passing for a 2xx response.
func HTTP(ctx context.Context, target HTTPTarget) Result {
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
