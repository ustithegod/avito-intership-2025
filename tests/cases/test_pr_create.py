import uuid

import pytest
import requests


def _make_team(session, base_url, team_name, members):
    return session.post(
        f"{base_url}/team/add", json={"team_name": team_name, "members": members}
    )


@pytest.mark.e2e
def test_create_pr_assign_2_reviewers(
    session: requests.Session, base_url: str, admin_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "A", "is_active": True},  # author
        {"user_id": "u2", "username": "B", "is_active": True},
        {"user_id": "u3", "username": "C", "is_active": True},
        {"user_id": "u4", "username": "D", "is_active": False},
    ]
    _make_team(session, base_url, team, members)
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    r = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Alpha",
            "author_id": "u1",
        },
    )
    assert r.status_code == 201, r.text
    assigned = r.json()["pr"]["assigned_reviewers"]
    assert len(assigned) == 2
    assert "u1" not in assigned


@pytest.mark.e2e
def test_create_pr_assign_1_reviewer(
    session: requests.Session, base_url: str, admin_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "A", "is_active": True},  # author
        {"user_id": "u2", "username": "B", "is_active": True},
        {"user_id": "u3", "username": "C", "is_active": False},
    ]
    _make_team(session, base_url, team, members)
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    r = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Beta",
            "author_id": "u1",
        },
    )
    assert r.status_code == 201, r.text
    assigned = r.json()["pr"]["assigned_reviewers"]
    assert len(assigned) == 1
    assert "u1" not in assigned


@pytest.mark.e2e
def test_create_pr_assign_0_reviewers(
    session: requests.Session, base_url: str, admin_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "A", "is_active": True},  # author
        {"user_id": "u2", "username": "B", "is_active": False},
    ]
    _make_team(session, base_url, team, members)
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    r = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Gamma",
            "author_id": "u1",
        },
    )
    assert r.status_code == 201, r.text
    assigned = r.json()["pr"]["assigned_reviewers"]
    assert len(assigned) == 0


@pytest.mark.e2e
@pytest.mark.negative
def test_create_pr_duplicate_and_validation(
    session: requests.Session, base_url: str, admin_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    _make_team(
        session, base_url, team, [{"user_id": "u1", "username": "A", "is_active": True}]
    )

    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    ok = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Delta",
            "author_id": "u1",
        },
    )
    assert ok.status_code == 201

    dup = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Delta",
            "author_id": "u1",
        },
    )
    assert dup.status_code == 409
    assert dup.json()["error"]["code"] == "PR_EXISTS"

    bad = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": "x",
            "pull_request_name": "abcd",
            "author_id": "u1",
        },
    )
    assert bad.status_code == 400
    assert bad.json()["error"]["code"] == "VALIDATION_ERROR"


@pytest.mark.e2e
@pytest.mark.negative
def test_create_pr_author_not_found(
    session: requests.Session, base_url: str, admin_headers: dict
):
    r = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": f"pr-{uuid.uuid4().hex[:8]}",
            "pull_request_name": "Feature Zeta",
            "author_id": "no-such",
        },
    )
    assert r.status_code == 404
