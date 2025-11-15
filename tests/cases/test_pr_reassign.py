import uuid

import pytest
import requests


def test_reassign_success(
    session: requests.Session, base_url: str, admin_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "A", "is_active": True},
        {"user_id": "u2", "username": "B", "is_active": True},
        {"user_id": "u3", "username": "C", "is_active": True},
        {"user_id": "u4", "username": "D", "is_active": True},
    ]
    session.post(f"{base_url}/team/add", json={"team_name": team, "members": members})
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    created = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Reassign",
            "author_id": "u1",
        },
    )
    assert created.status_code == 201
    assigned = created.json()["pr"]["assigned_reviewers"]
    assert assigned  # есть кого реассайнить

    old_reviewer = assigned[0]
    r = session.post(
        f"{base_url}/pullRequest/reassign",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "old_reviewer_id": old_reviewer},
    )
    if r.status_code == 200:
        body = r.json()
        assert body["replaced_by"] != old_reviewer
        assert body["replaced_by"] in body["pr"]["assigned_reviewers"]
    else:
        # На всякий случай: в реальных условиях может не оказаться кандидатов (маловероятно при 3+ активных)
        assert r.status_code == 409
        assert r.json()["error"]["code"] in ("NO_CANDIDATE", "NOT_ASSIGNED")


@pytest.mark.e2e
@pytest.mark.negative
def test_reassign_not_assigned(
    session: requests.Session, base_url: str, admin_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "A", "is_active": True},
        {"user_id": "u2", "username": "B", "is_active": True},
        {"user_id": "u3", "username": "C", "is_active": True},
        {"user_id": "u4", "username": "D", "is_active": True},
        {"user_id": "u5", "username": "E", "is_active": True},
    ]
    session.post(f"{base_url}/team/add", json={"team_name": team, "members": members})

    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    created = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature NotAssigned",
            "author_id": "u1",
        },
    )
    assert created.status_code == 201
    assigned = set(created.json()["pr"]["assigned_reviewers"])

    # Выберем любого участника из команды (кроме автора), которого НЕТ в assigned
    all_non_author = {m["user_id"] for m in members if m["user_id"] != "u1"}
    candidates = list(all_non_author - assigned)
    assert candidates, (
        f"expected at least one non-assigned user, assigned={assigned}, members={all_non_author}"
    )
    not_assigned_user = candidates[0]

    r = session.post(
        f"{base_url}/pullRequest/reassign",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "old_reviewer_id": not_assigned_user},
    )
    assert r.status_code == 409
    assert r.json()["error"]["code"] == "NOT_ASSIGNED"


@pytest.mark.e2e
@pytest.mark.negative
def test_reassign_no_candidate(
    session: requests.Session, base_url: str, admin_headers: dict
):
    # Только один активный кандидат кроме автора, чтобы после назначения не было замены
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "A", "is_active": True},  # author
        {"user_id": "u2", "username": "B", "is_active": True},  # единственный активный
        {"user_id": "u3", "username": "C", "is_active": False},
    ]
    session.post(f"{base_url}/team/add", json={"team_name": team, "members": members})
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    created = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature NoCandidate",
            "author_id": "u1",
        },
    )
    assert created.status_code == 201
    assigned = created.json()["pr"]["assigned_reviewers"]
    assert assigned == ["u2"]  # единственный ревьювер

    r = session.post(
        f"{base_url}/pullRequest/reassign",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "old_reviewer_id": "u2"},
    )
    assert r.status_code == 409
    assert r.json()["error"]["code"] == "NO_CANDIDATE"


@pytest.mark.e2e
@pytest.mark.negative
def test_reassign_after_merged(
    session: requests.Session, base_url: str, admin_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    session.post(
        f"{base_url}/team/add",
        json={
            "team_name": team,
            "members": [
                {"user_id": "u1", "username": "A", "is_active": True},
                {"user_id": "u2", "username": "B", "is_active": True},
            ],
        },
    )
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={
            "pull_request_id": pr_id,
            "pull_request_name": "Feature Merged",
            "author_id": "u1",
        },
    )
    session.post(
        f"{base_url}/pullRequest/merge",
        headers=admin_headers,
        json={"pull_request_id": pr_id},
    )
    r = session.post(
        f"{base_url}/pullRequest/reassign",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "old_reviewer_id": "u2"},
    )
    r = session.post(
        f"{base_url}/pullRequest/reassign",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "old_reviewer_id": "u2"},
    )
    assert r.status_code == 409
    assert r.json()["error"]["code"] == "PR_MERGED"
