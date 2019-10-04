package trace

import (
	"context"
	"testing"

	. "go.opencensus.io/trace"
)

var (
	tid = TraceID{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 4, 8, 16, 32, 64, 128}
	sid = SpanID{1, 2, 4, 8, 16, 32, 64, 128}
)

func TestRespectParentSampler(t *testing.T) {
	for _, test := range []struct {
		remoteParent       bool
		localParent        bool
		parentTraceOptions TraceOptions
		fallbackSampler    Sampler
		wantTraceOptions   TraceOptions
	}{
		// Remote parents
		{true, false, 0, NeverSample(), 0},
		{true, false, 1, NeverSample(), 1},
		{true, false, 0, AlwaysSample(), 0},
		{true, false, 1, AlwaysSample(), 1},

		// Local parents
		{false, true, 0, NeverSample(), 0},
		{false, true, 1, NeverSample(), 1},
		{false, true, 0, AlwaysSample(), 0},
		{false, true, 1, AlwaysSample(), 1},

		// No parents
		{false, false, 0, NeverSample(), 0},
		{false, false, 0, AlwaysSample(), 1},
	} {
		sampler := RespectParentSampler(test.fallbackSampler)

		var ctx context.Context
		if test.remoteParent {
			sc := SpanContext{
				TraceID:      tid,
				SpanID:       sid,
				TraceOptions: test.parentTraceOptions,
			}
			ctx, _ = StartSpanWithRemoteParent(context.Background(), "foo", sc, WithSampler(sampler))
		} else if test.localParent {
			sampler := NeverSample()
			if test.parentTraceOptions == 1 {
				sampler = AlwaysSample()
			}
			ctx2, _ := StartSpan(context.Background(), "foo", WithSampler(sampler))
			ctx, _ = StartSpan(ctx2, "foo", WithSampler(sampler))
		} else {
			ctx, _ = StartSpan(context.Background(), "foo", WithSampler(sampler))
		}

		sc := FromContext(ctx).SpanContext()
		if (sc == SpanContext{}) {
			t.Errorf("case %#v: starting new span: no span in context", test)
			continue
		}
		if sc.SpanID == (SpanID{}) {
			t.Errorf("case %#v: starting new span: got zero SpanID, want nonzero", test)
		}
		if sc.TraceOptions != test.wantTraceOptions {
			t.Errorf("case %#v: starting new span: got TraceOptions %x, want %x", test, sc.TraceOptions, test.wantTraceOptions)
		}
	}
}
