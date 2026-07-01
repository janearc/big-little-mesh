# frood.emit -- relay a bento lifecycle event to the local Go sidecar, which
# owns the kafka / schema-registry / protobuf wire. Python never touches kafka.
#
# Best-effort: emit() never raises. A sidecar or bus hiccup must not break a bento --
# the bus is the durable record, and a missed emit is recovered by re-handling the
# work, never by failing it. Modeled on paling's daemon emit (stdlib urllib, no deps
# beyond protobuf).
import logging
import os
import urllib.request
import uuid

from google.protobuf import json_format

from bento.v1 import bento_pb2

log = logging.getLogger(__name__)

# the local Go sidecar's intake. Override with FROOD_SIDECAR_URL.
_SIDECAR_URL = os.environ.get("FROOD_SIDECAR_URL", "http://localhost:9090/emit")


def emit(event, sidecar_url=None) -> bool:
    # POST the event as protojson (the contract's canonical JSON) to the sidecar. Returns True iff
    # the sidecar ACCEPTED it (a 2xx) -- and for a TERMINAL event a 2xx means the bus broker acked
    # it durably, because the sidecar refuses (non-2xx) a terminal it could not publish. Never
    # raises: any failure (sidecar down, a 4xx/5xx, a timeout) is logged and returns False, so a
    # best-effort caller can ignore the result while a terminal caller can gate delivery on it.
    url = sidecar_url or _SIDECAR_URL
    try:
        body = json_format.MessageToJson(event).encode()
        req = urllib.request.Request(
            url, data=body, headers={"Content-Type": "application/json"}, method="POST"
        )
        urllib.request.urlopen(req, timeout=2).close()
        return True
    except Exception as e:  # noqa: BLE001 - emit never breaks the bento; the caller decides on False
        log.warning("frood: emit to sidecar failed: %s", e)
        return False


# TERMINAL bento states: the FSM halts here (DONE/PARTIAL/FAILED), so the event is the bento's
# output commit that must reach the bus with a broker ack. There is no generated terminal-set (the
# FSM models terminal as "a handler returns UNSPECIFIED"), so it is named here; keep it in sync
# with bento.proto's terminal states (and cmd/sidecar's isTerminal).
_TERMINAL_STATES = frozenset(
    {
        bento_pb2.BENTO_STATE_DONE,
        bento_pb2.BENTO_STATE_PARTIAL,
        bento_pb2.BENTO_STATE_FAILED,
    }
)


def sidecar_emitter(sidecar_url=None, handler=None, ack_holder=None):
    # build an Emitter for the generated fsm harness: step() calls emitter(b, state) on each
    # transition, and this relays a BentoLifecycleEvent to the sidecar. event_id is a fresh
    # uuid4 (the idempotency key); the bento carries its own id and kind.
    #
    # the failure REASON lives on the HANDLER, not the bento, so the emitter closes over the
    # handler to populate error_message on a FAILED event -- without this, a FAILED bento's
    # bus event carried no error and there was no provenance/correlation. trace_id correlates
    # one bento's processing; the bento has no trace field yet, so trace_id == bento_id for
    # now (a bento's events share its id as the trace until a real trace field lands).
    #
    # ack_holder (optional): a dict the emitter records the TERMINAL event's ack into
    # (ack_holder["terminal_acked"] = True/False), so the caller can gate delivery on the bento's
    # commit actually landing on the bus. Non-terminal emits stay best-effort -- their result is
    # not recorded and a drop is recovered by re-running the bento.
    def _emit(b, state):
        ev = bento_pb2.BentoLifecycleEvent(
            event_id=str(uuid.uuid4()),
            trace_id=b.id,
            bento_id=b.id,
            bento_kind=b.kind,
            state=state,
            handler=type(handler).__name__ if handler is not None else "",
        )
        if state == bento_pb2.BENTO_STATE_FAILED and handler is not None:
            ev.error_message = getattr(handler, "error", "") or ""
        acked = emit(ev, sidecar_url)
        if state in _TERMINAL_STATES and ack_holder is not None:
            ack_holder["terminal_acked"] = acked

    return _emit
