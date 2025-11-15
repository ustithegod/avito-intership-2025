import uuid

import pytest
import requests


@pytest.mark.auth
def test_no_token_user_endpoints(session: requests.Session, base_url: str):
    r = session.get(f"{base_url}/team/get", params={"team_name": "any"})
    assert r.status_code == 401

    r = session.get(f"{base_url}/users/getReview", params={"user_id": "u1"})
    assert r.status_code == 401


@pytest.mark.auth
def test_no_token_admin_endpoints(session: requests.Session, base_url: str):
    r = session.post(
        f"{base_url}/users/setIsActive", json={"user_id": "u1", "is_active": False}
    )
    assert r.status_code == 401

    r = session.post(
        f"{base_url}/pullRequest/create",
        json={"pull_request_id": "x", "pull_request_name": "abcde", "author_id": "u1"},
    )
    assert r.status_code == 401

    r = session.post(f"{base_url}/pullRequest/merge", json={"pull_request_id": "x"})
    assert r.status_code == 401

    r = session.post(
        f"{base_url}/pullRequest/reassign",
        json={"pull_request_id": "x", "old_reviewer_id": "u2"},
    )
    assert r.status_code == 401


@pytest.mark.auth
def test_user_token_on_admin_endpoints(
    session: requests.Session, base_url: str, user_headers: dict
):
    r = session.post(
        f"{base_url}/users/setIsActive",
        headers=user_headers,
        json={"user_id": "u1", "is_active": False},
    )
    assert r.status_code == 401

    r = session.post(
        f"{base_url}/pullRequest/create",
        headers=user_headers,
        json={"pull_request_id": "x", "pull_request_name": "abcde", "author_id": "u1"},
    )
    assert r.status_code == 401


@pytest.mark.auth
def test_admin_token_on_user_endpoints(
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
    r = session.get(
        f"{base_url}/team/get", headers=admin_headers, params={"team_name": team}
    )
    assert r.status_code == 200


@pytest.mark.auth
def test_invalid_token(session: requests.Session, base_url: str):
    r = session.get(
        f"{base_url}/team/get",
        headers={"Authorization": "Bearer invalid"},
        params={"team_name": "any"},
    )
    assert r.status_code == 401
