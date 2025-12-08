import os
import sys
from dotenv import load_dotenv
from dataclasses import dataclass
from typing import List


@dataclass
class SenderConfig:
    max_retries: int = 3
    next_retry_delay: int = 2
    timeout: int = 10


@dataclass
class AuthConfig:
    client_id: str
    tenant: str
    email: str
    scopes: List[str]
    auth_method: str


def load_auth_config() -> AuthConfig:
    load_dotenv()

    cfg = AuthConfig(
        client_id=get_env("CLIENT_ID", ""),
        tenant=get_env("TENANT_ID", ""),
        email=get_env("EMAIL", ""),
        scopes=get_env(
            "SCOPES",
            "https://graph.microsoft.com/.default"
        ).split(","),
        auth_method=get_env("AUTH_METHOD", "DEVICE_CODE"),
    )

    validate(cfg)
    return cfg


def get_env(key: str, fallback: str) -> str:
    return os.getenv(key, fallback)


def validate(cfg: AuthConfig):
    if not cfg.client_id:
        print("Missing CLIENT ID", file=sys.stderr)
        sys.exit(1)

    if not cfg.tenant:
        print("Missing TENANT ID", file=sys.stderr)
        sys.exit(1)

    if not cfg.email:
        print("Missing EMAIL", file=sys.stderr)
        sys.exit(1)

    if cfg.auth_method not in ("DEVICE_CODE", "INTERACTIVE"):
        print("AUTH METHOD must be either DEVICE_CODE or INTERACTIVE", file=sys.stderr)
        sys.exit(1)
