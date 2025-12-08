import json
import subprocess
import platform
import pathlib
from typing import Any

import config


class TeamsClient:
    def __init__(self):
        self.proc = subprocess.Popen(
            [str(self._binary())],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            bufsize=1,
        )

        self.init_client()

    def get_response(self) -> Any:
        res = json.loads(self.proc.stdout.readline())

        if "error" in res and res["error"]:
            raise RuntimeError(res["error"])

        return res["result"]

    def _binary(self):
        base = pathlib.Path(__file__).parent / "bin"
        osname = platform.system()

        if osname == "Windows":
            return base / "APIbridge_windows.exe"
        elif osname == "Linux":
            return base / "APIbridge_linux"
        else:
            raise RuntimeError("Unsupported OS")

    def init_client(self) -> Any:
        sender_config = config.SenderConfig()
        auth_config = config.load_auth_config()
        self.proc.stdin.write(
            json.dumps(
                {
                    "type": "init",
                    "config": {
                        "senderConfig": {
                            "maxRetries": sender_config.max_retries,
                            "nextRetryDelay": sender_config.next_retry_delay,
                            "timeout": sender_config.timeout,
                        },
                        "authConfig": {
                            "clientId": auth_config.client_id,
                            "tenant": auth_config.tenant,
                            "email": auth_config.email,
                            "scopes": auth_config.scopes,
                            "authMethod": auth_config.auth_method,
                        },
                    },
                }
            )
            + "\n"
        )
        self.proc.stdin.flush()
        self.get_response()



    def list_channels(self, teamRef: str) -> Any:
        req = {"type": "request", "method": "listChannels", "params": {"teamRef": teamRef}}

        self.proc.stdin.write(json.dumps(req) + "\n")
        self.proc.stdin.flush()

        return self.get_response()
