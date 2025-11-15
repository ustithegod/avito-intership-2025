import uuid

import pytest
import requests


def test_statistics_basic_and_sort(
    session: requests.Session, base_url: str, user_headers: dict, admin_headers: dict
):
    # подготовим немного данных
    team = f"team-{uuid.uuid4().hex[:8]}"
    session.post(
        f"{base_url}/team/add",
        json={
            "team_name": team,
            "members": [{"user_id": "u1", "username": "A", "is_active": True}],
        },
    )
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Stats",
            "author_id": "u1",
        },
    )

    # корректные запросы
    r1 = session.get(f"{base_url}/statistics", headers=user_headers)
    assert r1.status_code == 200
    data = r1.json()
    assert "pr" in data and "users" in data

    r2 = session.get(
        f"{base_url}/statistics", headers=admin_headers, params={"sort": "asc"}
    )
    assert r2.status_code == 200

    # неверный sort
    bad = session.get(
        f"{base_url}/statistics", headers=user_headers, params={"sort": "zzz"}
    )
    bad = session.get(f"{base_url}/statistics", headers=user_headers, params={"sort": "zzz"})
    assert bad.status_code == 400
