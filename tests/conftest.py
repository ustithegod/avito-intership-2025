import os
import time
import typing as t
from datetime import UTC, datetime, timedelta

import jwt
import pytest
import requests

DEFAULT_BASE_URL = os.getenv("TEST_BASE_URL", "http://localhost:8081")
ADMIN_SECRET = os.getenv("ADMIN_JWT_SECRET", "admin_secret_key")
USER_SECRET = os.getenv("USER_JWT_SECRET", "user_secret_key")


def _make_token(secret: str, role: str, expires_in_minutes: int = 60 * 24) -> str:
    now = datetime.now(UTC)
    payload = {
        "role": role,
        "exp": int((now + timedelta(minutes=expires_in_minutes)).timestamp()),
        "iat": int(now.timestamp()),
    }
    return jwt.encode(payload, secret, algorithm="HS256")


@pytest.fixture(scope="session")
def base_url() -> str:
    return DEFAULT_BASE_URL.rstrip("/")


@pytest.fixture(scope="session")
def session() -> t.Iterator[requests.Session]:
    s = requests.Session()
    yield s
    s.close()


@pytest.fixture(scope="session", autouse=True)
def wait_for_app(session: requests.Session, base_url: str):
    # Ждем готовности сервиса
    deadline = time.time() + 60
    last_exc = None
    while time.time() < deadline:
        try:
            r = session.get(f"{base_url}/health", timeout=1.5)
            if r.status_code == 200:
                return
        except Exception as e:
            last_exc = e
        time.sleep(0.5)
    raise RuntimeError(f"Service at {base_url} is not healthy. Last error: {last_exc}")


@pytest.fixture(scope="session")
def admin_token() -> str:
    return _make_token(ADMIN_SECRET, "admin")


@pytest.fixture(scope="session")
def user_token() -> str:
    return _make_token(USER_SECRET, "user")


@pytest.fixture()
def admin_headers(admin_token: str) -> dict:
    return {"Authorization": f"Bearer {admin_token}"}


@pytest.fixture()
def user_headers(user_token: str) -> dict:
    return {"Authorization": f"Bearer {user_token}"}
