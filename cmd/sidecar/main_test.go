package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"

	bentov1 "github.com/janearc/big-little-mesh/gen/go/bento/v1"
)

// TestIsTerminal pins the terminal-outcome set (DONE/PARTIAL/FAILED); the FSM's active states
// NOTICED/COOK and UNSPECIFIED are not terminal.
func TestIsTerminal(t *testing.T) {
	for _, s := range []bentov1.BentoState{
		bentov1.BentoState_BENTO_STATE_DONE,
		bentov1.BentoState_BENTO_STATE_PARTIAL,
		bentov1.BentoState_BENTO_STATE_FAILED,
	} {
		if !isTerminal(s) {
			t.Errorf("%v should be terminal", s)
		}
	}
	for _, s := range []bentov1.BentoState{
		bentov1.BentoState_BENTO_STATE_UNSPECIFIED,
		bentov1.BentoState_BENTO_STATE_NOTICED,
		bentov1.BentoState_BENTO_STATE_COOK,
	} {
		if isTerminal(s) {
			t.Errorf("%v should not be terminal", s)
		}
	}
}

// TestEmitIntake_DryBus pins the terminal-commit contract on a dry bus (nil publisher): a terminal
// event is REFUSED (503) so the producer does not record the work as delivered, while a
// non-terminal event is accept-and-dropped best-effort (202).
func TestEmitIntake_DryBus(t *testing.T) {
	h := emitIntake(context.Background(), nil) // nil publisher == dry / unreachable bus

	body := func(state bentov1.BentoState) string {
		b, err := protojson.Marshal(&bentov1.BentoLifecycleEvent{BentoId: "b1", State: state})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	for _, tc := range []struct {
		name  string
		state bentov1.BentoState
		want  int
	}{
		{"terminal DONE refused", bentov1.BentoState_BENTO_STATE_DONE, http.StatusServiceUnavailable},
		{"terminal PARTIAL refused", bentov1.BentoState_BENTO_STATE_PARTIAL, http.StatusServiceUnavailable},
		{"terminal FAILED refused", bentov1.BentoState_BENTO_STATE_FAILED, http.StatusServiceUnavailable},
		{"non-terminal NOTICED accepted", bentov1.BentoState_BENTO_STATE_NOTICED, http.StatusAccepted},
		{"non-terminal COOK accepted", bentov1.BentoState_BENTO_STATE_COOK, http.StatusAccepted},
	} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/emit", strings.NewReader(body(tc.state)))
		h(rr, req)
		if rr.Code != tc.want {
			t.Errorf("%s: code = %d, want %d", tc.name, rr.Code, tc.want)
		}
	}
}
