// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package diagnostics

import (
	"context"
	"reflect"
	"testing"

	"github.com/dapr/dapr/pkg/config"
	"github.com/valyala/fasthttp"
	"go.opencensus.io/trace"
)

var (
	tpHeader = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	traceID  = trace.TraceID{75, 249, 47, 53, 119, 179, 77, 166, 163, 206, 146, 157, 14, 14, 71, 54}
	spanID   = trace.SpanID{0, 240, 103, 170, 11, 169, 2, 183}
	traceOpt = trace.TraceOptions(1)
)

func TestStartClientSpanTracing(t *testing.T) {
	req := getTestHTTPRequest()
	spec := config.TracingSpec{SamplingRate: "0.5"}

	StartTracingClientSpanFromHTTPContext(context.Background(), req, "test", spec)
}

func TestSpanContextFromRequest(t *testing.T) {
	tests := []struct {
		name   string
		header string
		wantSc trace.SpanContext
		wantOk bool
	}{
		{
			name:   "future version",
			header: "02-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{75, 249, 47, 53, 119, 179, 77, 166, 163, 206, 146, 157, 14, 14, 71, 54},
				SpanID:       trace.SpanID{0, 240, 103, 170, 11, 169, 2, 183},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name:   "zero trace ID and span ID",
			header: "00-00000000000000000000000000000000-0000000000000000-01",
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name:   "valid header",
			header: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
			wantSc: trace.SpanContext{
				TraceID:      trace.TraceID{75, 249, 47, 53, 119, 179, 77, 166, 163, 206, 146, 157, 14, 14, 71, 54},
				SpanID:       trace.SpanID{0, 240, 103, 170, 11, 169, 2, 183},
				TraceOptions: trace.TraceOptions(1),
			},
			wantOk: true,
		},
		{
			name:   "missing options",
			header: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7",
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
		{
			name:   "empty options",
			header: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-",
			wantSc: trace.SpanContext{},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &fasthttp.Request{}
			req.Header.Add("traceparent", tt.header)

			gotSc, gotOk := SpanContextFromRequest(req)
			if !reflect.DeepEqual(gotSc, tt.wantSc) {
				t.Errorf("SpanContextFromRequest() gotSc = %v, want %v", gotSc, tt.wantSc)
			}
			if gotOk != tt.wantOk {
				t.Errorf("SpanContextFromRequest gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestSpanContextToRequest(t *testing.T) {
	tests := []struct {
		sc         trace.SpanContext
		wantHeader string
	}{
		{
			sc: trace.SpanContext{
				TraceID:      trace.TraceID{75, 249, 47, 53, 119, 179, 77, 166, 163, 206, 146, 157, 14, 14, 71, 54},
				SpanID:       trace.SpanID{0, 240, 103, 170, 11, 169, 2, 183},
				TraceOptions: trace.TraceOptions(1),
			},
			wantHeader: "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.wantHeader, func(t *testing.T) {
			req := &fasthttp.Request{}
			SpanContextToRequest(tt.sc, req)

			h := string(req.Header.Peek("traceparent"))
			if got, want := h, tt.wantHeader; got != want {
				t.Errorf("SpanContextToRequest() header = %v, want %v", got, want)
			}
		})
	}
}

func getTestHTTPRequest() *fasthttp.Request {
	req := &fasthttp.Request{}
	req.Header.Set("dapr-testheaderkey", "dapr-testheadervalue")
	req.Header.Set("x-testheaderkey1", "dapr-testheadervalue")
	req.Header.Set("daprd-testheaderkey2", "dapr-testheadervalue")

	var (
		tid = trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 4, 8, 16, 32, 64, 128}
		sid = trace.SpanID{1, 2, 4, 8, 16, 32, 64, 128}
	)

	sc := trace.SpanContext{
		TraceID:      tid,
		SpanID:       sid,
		TraceOptions: 0x0,
	}

	corID := SerializeSpanContext(sc)
	req.Header.Set(CorrelationID, corID)

	return req
}
