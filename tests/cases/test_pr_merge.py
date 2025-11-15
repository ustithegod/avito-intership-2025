import uuid

import pytest
import requests


@pytest.mark.e2e
def test_merge_idempotent(
    session: requests.Session, base_url: str, admin_headers: dict
):
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
            "pull_request_name": "Feature Merge",
            "author_id": "u1",
        },
    )
    r1 = session.post(
        f"{base_url}/pullRequest/merge",
        headers=admin_headers,
        json={"pull_request_id": pr_id},
    )
    assert r1.status_code == 200
    assert r1.json()["pr"]["status"] == "MERGED"
    r2 = session.post(
        f"{base_url}/pullRequest/merge",
        headers=admin_headers,
        json={"pull_request_id": pr_id},
    )
    assert r2.status_code == 200
    assert r2.json()["pr"]["status"] == "MERGED"


@pytest.mark.e2e
@pytest.mark.negative
def test_merge_not_found(session: requests.Session, base_url: str, admin_headers: dict):
    r = session.post(
        f"{base_url}/pullRequest/merge",
        headers=admin_headers,
        json={"pull_request_id": "no-such"},
    )
    assert r.status_code == 404
