from __future__ import annotations

import json
import os
import sys

import requests


def main() -> int:
    raw = sys.stdin.read()
    if not raw.strip():
        raise RuntimeError("Expected webhook envelope JSON on stdin")

    envelope = json.loads(raw)
    payload = envelope.get("payload") or {}
    headers = envelope.get("headers") or {}
    webhook_url = os.environ.get("WEBHOOK_BRIDGE_WECHAT_WEBHOOK_URL")
    if not webhook_url:
        raise RuntimeError("WEBHOOK_BRIDGE_WECHAT_WEBHOOK_URL is not set")

    repository = (payload.get("repository") or {}).get("full_name", "unknown")
    sender = (payload.get("sender") or {}).get("login", "unknown")
    event = headers.get("x-github-event") or headers.get("X-GitHub-Event") or "unknown"
    delivery = headers.get("x-github-delivery") or headers.get("X-GitHub-Delivery") or envelope.get("request_id")
    ref = payload.get("ref", "")

    text = "\n".join(
        [
            "Webhook Bridge received GitHub event",
            f"repo: {repository}",
            f"event: {event}",
            f"ref: {ref}",
            f"sender: {sender}",
            f"delivery: {delivery}",
        ]
    )

    response = requests.post(
        webhook_url,
        json={"msgtype": "text", "text": {"content": text}},
        timeout=10,
    )
    response.raise_for_status()
    try:
        response_body = response.json()
    except ValueError:
        response_body = {"body": response.text}

    print(
        json.dumps(
            {
                "status": "sent",
                "provider": "wechat",
                "repository": repository,
                "event": event,
                "delivery": delivery,
                "response": response_body,
            },
            ensure_ascii=False,
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
