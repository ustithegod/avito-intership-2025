import uuid

import pytest
import requests


@pytest.mark.e2e
def test_full_happy_flow(
    session: requests.Session, base_url: str, admin_headers: dict, user_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "A", "is_active": True},
        {"user_id": "u2", "username": "B", "is_active": True},
        {"user_id": "u3", "username": "C", "is_active": False},
    ]
    session.post(f"{base_url}/team/add", json={"team_name": team, "members": members})

    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    created = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Flow",
            "author_id": "u1",
        },
    )
    assert created.status_code == 201

    # reviewer list может быть 0..2, проверим только инварианты
    assigned = created.json()["pr"]["assigned_reviewers"]
    assert "u1" not in assigned

    # getReview
    if assigned:
        r = session.get(
            f"{base_url}/users/getReview",
            headers=user_headers,
            params={"user_id": assigned[0]},
        )
        assert r.status_code in (200, 404)

    # merge + повторно
    m1 = session.post(
        f"{base_url}/pullRequest/merge",
        headers=admin_headers,
        json={"pull_request_id": pr_id},
    )
    assert m1.status_code == 200
    m2 = session.post(
        f"{base_url}/pullRequest/merge",
        headers=admin_headers,
        json={"pull_request_id": pr_id},
    )
    assert m2.status_code == 200
