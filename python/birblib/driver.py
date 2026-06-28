# birblib.driver -- the standalone walk: drive a NOTICED bento to a terminal state.
#
# in production the bus is the loop -- one step per consumed event, a process may die
# between steps and another instance resumes from the bus (good_citizen.run_step). this
# is the LOCAL counterpart, for a CLI run with no bus: step the FSM until a terminal
# handler is reached, relaying each transition to an optional emitter. it is the magpie
# process() loop, lifted so no birb re-writes it.

from bento.v1 import bento_pb2
from good_citizen import fsm

from birblib.bento import BirbBento

_TERMINAL = {bento_pb2.BENTO_STATE_DONE, bento_pb2.BENTO_STATE_FAILED}


def run(handlers, bento, emitter=None) -> dict:
    # drive `bento` (a BirbBento or a bare bento_pb2.Bento) to a terminal state through
    # the generated FSM, relaying each transition to `emitter` when given (the CLI passes
    # None -- local, no bus). returns the handler's manifest (where the outputs are), NOT
    # the bytes. raises RuntimeError if the bento ends FAILED, so a caller surfaces the
    # error rather than reporting success.
    pb = bento.pb if isinstance(bento, BirbBento) else bento
    while pb.state not in _TERMINAL:
        prev = pb.state
        fsm.step(handlers, emitter, pb)
        if pb.state == prev:  # a handler that did not advance -- stop rather than spin
            break
    if pb.state == bento_pb2.BENTO_STATE_FAILED:
        raise RuntimeError(handlers.error or "birb bento failed")
    return handlers.manifest
