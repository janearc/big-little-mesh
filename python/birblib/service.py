# birblib.service -- the service scaffold a birb gets for free: the inbox daemon, the
# async-submit HTTP surface, and the CLI ACK shape. the per-modality /v1 facade (images
# vs audio vs video differ) stays the birb's own; the job API, health, and artifact
# serving are the lib's.
#
# fastapi/uvicorn are an OPTIONAL extra (`blm-good-citizen[service]`): they are imported
# lazily inside the HTTP functions, so a birb that only wants the bento/handlers/daemon
# never pays for them. the daemon half rides on good_citizen.watcher (stdlib) and needs no
# extra.

import json
import logging
import threading
from pathlib import Path

from bento.v1 import bento_pb2
from good_citizen import fsm, watcher
from good_citizen import emit as _emit

from birblib.bento import BirbBento

logger = logging.getLogger(__name__)

_TERMINAL = {bento_pb2.BENTO_STATE_DONE, bento_pb2.BENTO_STATE_FAILED}


# --- driving a bento with the REAL sidecar emit (the daemon/http path) ------------

def drive(handlers, bento, sidecar_url=None) -> dict | None:
    # step a bento to a terminal state, relaying each transition to the Go sidecar (the
    # bus). unlike driver.run (the CLI, which raises on FAILED so a shell sees the error),
    # this returns the manifest and never raises on a FAILED bento -- a daemon logs the
    # outcome and stays up for the next file. an unexpected exception still propagates so
    # the watcher's per-file guard can catch it.
    pb = bento.pb if isinstance(bento, BirbBento) else bento
    emitter = _emit.sidecar_emitter(sidecar_url)
    while pb.state not in _TERMINAL:
        prev = pb.state
        fsm.step(handlers, emitter, pb)
        if pb.state == prev:
            break
    return handlers.manifest


# --- the inbox daemon -------------------------------------------------------------

def serve_inbox(
    inbox,
    make_handlers,
    make_bento,
    *,
    suffixes=None,
    sidecar_url=None,
    interval=5.0,
) -> None:
    # watch `inbox`, and for each new file build a bento (make_bento(path) -> BirbBento in
    # NOTICED) and drive it to terminal with the real sidecar emit. make_handlers() -> a
    # fresh BirbHandlers per file. dup-over-loss is preserved: the watcher never moves or
    # deletes the source, and on_noticed COPIES it into the bento.
    def _handle(path):
        bento = make_bento(Path(path))
        manifest = drive(make_handlers(), bento, sidecar_url=sidecar_url)
        if manifest is not None:
            logger.info(
                "birblib: %s -> %s (ok=%s)",
                Path(path).name, manifest.get("artifact"), manifest.get("ok"),
            )

    watcher.watch(inbox, _handle, suffixes=suffixes, interval=interval)


# --- the HTTP surface (async submit + poll + artifact serving) --------------------

def read_manifest(bentos_root, bento_id: str) -> dict | None:
    # the on-disk manifest for a bento, or None if it has not been written yet (the job is
    # still running) -- the manifest IS the job record on the local path.
    path = Path(bentos_root) / bento_id / "manifest.json"
    return json.loads(path.read_text()) if path.is_file() else None


def _safe_artifact_path(bentos_root, bento_id: str, name: str) -> Path | None:
    # resolve an artifact request to a path UNDER the bento's outputs dir, or None if the
    # name would escape it (the traversal guard). a crafted "../../etc/passwd" resolves
    # outside `base` and is rejected; only a path that stays within outputs is served.
    base = (Path(bentos_root) / bento_id / "outputs").resolve()
    target = (base / name).resolve()
    if base != target and base not in target.parents:
        return None
    return target


def build_app(*, name, bentos_root, make_handlers, make_bento, sidecar_url=None):
    # build the birb's HTTP app: /health, 202 submit (POST /jobs) + poll (GET /jobs/{id}),
    # and artifact serving with the traversal guard. submit NEVER renders inline -- it
    # builds the bento (so the id is known), returns 202 with that id, and drives the work
    # on a background thread; the caller polls. the in-memory thread is the local runner;
    # the durable record is the on-disk bento + manifest (and the bus, via the emit).
    from fastapi import FastAPI, HTTPException
    from fastapi.responses import FileResponse

    app = FastAPI(title=name)

    @app.get("/health")
    def health():
        return {"status": "ok", "service": name}

    @app.post("/jobs", status_code=202)
    def submit(request: dict):
        bento = make_bento(request)
        bento_id = bento.pb.id
        handlers = make_handlers()

        def _run():
            try:
                drive(handlers, bento, sidecar_url=sidecar_url)
            except Exception as e:  # noqa: BLE001 - one job's failure must not kill the worker
                logger.error("birblib: job %s crashed: %s", bento_id, e)

        threading.Thread(target=_run, daemon=True).start()
        return {"status": "accepted", "bento_id": bento_id, "job": f"/jobs/{bento_id}"}

    @app.get("/jobs/{bento_id}")
    def job(bento_id: str):
        manifest = read_manifest(bentos_root, bento_id)
        if manifest is None:
            if (Path(bentos_root) / bento_id).is_dir():
                return {"status": "running", "bento_id": bento_id}
            raise HTTPException(status_code=404, detail="no such job")
        return {
            "status": "done" if manifest["ok"] else manifest["state"].lower(),
            "manifest": manifest,
        }

    @app.get("/artifacts/{bento_id}/{name:path}")
    def artifact(bento_id: str, name: str):
        path = _safe_artifact_path(bentos_root, bento_id, name)
        if path is None or not path.is_file():
            raise HTTPException(status_code=404, detail="no such artifact")
        return FileResponse(path)

    return app


# --- the CLI ACK ------------------------------------------------------------------

def ack(manifest: dict, message: str = "") -> dict:
    # the JSON a birb's CLI prints: an ACK of where the result landed, not the bytes. it
    # is json-by-default (a birb is a good agent-citizen) and reports ok + the artifact
    # path, so an agent caller can find the output and know whether it worked.
    out = {
        "status": "ok" if manifest.get("ok") else "incomplete",
        "bento_id": manifest.get("bento_id"),
        "state": manifest.get("state"),
        "artifact": manifest.get("artifact"),
    }
    if message:
        out["message"] = message
    return out
