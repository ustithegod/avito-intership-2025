import uuid
import pytest
import requests


@pytest.mark.e2e
def test_get_review_success_for_assigned_reviewer_user_role(session: requests.Session, base_url: str, admin_headers: dict, user_headers: dict):
    team = f"team-{uuid.uuid4().hex[:8]}"
    members = [
        {"user_id": "u1", "username": "Alice", "is_active": True},   # author
        {"user_id": "u2", "username": "Bob", "is_active": True},
        {"user_id": "u3", "username": "Carol", "is_active": True},
        {"user_id": "u4", "username": "Dave", "is_active": False},
    ]
    session.post(f"{base_url}/team/add", json={"team_name": team, "members": members})

    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    r = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "pull_request_name": "Feature GetReview", "author_id": "u1"},
    )
    assert r.status_code == 201, r.text
    assigned = r.json()["pr"]["assigned_reviewers"]

    if not assigned:
        pytest.skip("No reviewers were assigned due to team composition; skip success check")

    reviewer = assigned[0]
    gr = session.get(f"{base_url}/users/getReview", headers=user_headers, params={"user_id": reviewer})
    assert gr.status_code == 200, gr.text
    body = gr.json()
    assert body["user_id"] == reviewer
    # PR может быть 1 или более; проверим, что наш pr_id там присутствует
    assert any(p["pull_request_id"] == pr_id for p in body["pull_requests"])


@pytest.mark.e2e
def test_get_review_success_with_admin_token(session: requests.Session, base_url: str, admin_headers: dict):
    team = f"team-{uuid.uuid4().hex[:8]}"
    session.post(f"{base_url}/team/add", json={
        "team_name": team,
        "members": [
            {"user_id": "u1", "username": "Alice", "is_active": True},
            {"user_id": "u2", "username": "Bob", "is_active": True},
        ],
    })
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    created = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "pull_request_name": "Feature AdminToken", "author_id": "u1"},
    )
    assert created.status_code == 201
    assigned = created.json()["pr"]["assigned_reviewers"]
    if not assigned:
        pytest.skip("No reviewers assigned; skip admin-token getReview check")

    reviewer = assigned[0]
    gr = session.get(f"{base_url}/users/getReview", headers=admin_headers, params={"user_id": reviewer})
    assert gr.status_code == 200


@pytest.mark.e2e
@pytest.mark.negative
def test_get_review_not_found(session: requests.Session, base_url: str, user_headers: dict):
    r = session.get(f"{base_url}/users/getReview", headers=user_headers, params={"user_id": "no-such-user"})
    assert r.status_code == 404


@pytest.mark.e2e
@pytest.mark.negative
def test_get_review_missing_param_400(session: requests.Session, base_url: str, user_headers: dict):
    r = session.get(f"{base_url}/users/getReview", headers=user_headers)
    assert r.status_code == 400
    assert r.json()["error"]["code"] == "BAD_REQUEST"


@pytest.mark.e2e
def test_set_is_active_success_affects_assignment(session: requests.Session, base_url: str, admin_headers: dict):
    team = f"team-{uuid.uuid4().hex[:8]}"
    # u2 активен -> потом деактивируем и убедимся, что он не назначается ревьювером
    session.post(f"{base_url}/team/add", json={
        "team_name": team,
        "members": [
            {"user_id": "u1", "username": "Alice", "is_active": True},   # author
            {"user_id": "u2", "username": "Bob", "is_active": True},
            {"user_id": "u3", "username": "Carol", "is_active": True},
        ],
    })

    # Деактивируем u2
    r = session.post(f"{base_url}/users/setIsActive", headers=admin_headers, json={"user_id": "u2", "is_active": False})
    assert r.status_code == 200
    assert r.json()["user"]["user_id"] == "u2"
    assert r.json()["user"]["is_active"] is False

    # Создадим PR и убедимся, что u2 не назначен
    pr_id = f"pr-{uuid.uuid4().hex[:8]}"
    created = session.post(
        f"{base_url}/pullRequest/create",
        headers=admin_headers,
        json={"pull_request_id": pr_id, "pull_request_name": "Feature InactiveU2", "author_id": "u1"},
    )
    assert created.status_code == 201, created.text
    assigned = created.json()["pr"]["assigned_reviewers"]
    assert "u2" not in assigned


@pytest.mark.e2e
@pytest.mark.negative
def test_set_is_active_nonexistent_user_404(session: requests.Session, base_url: str, admin_headers: dict):
    r = session.post(f"{base_url}/users/setIsActive", headers=admin_headers, json={"user_id": "no-such-user", "is_active": False})
    assert r.status_code == 404
    assert r.json()["error"]["code"] == "NOT_FOUND"


@pytest.mark.e2e
@pytest.mark.auth
def test_set_is_active_requires_admin_token(session: requests.Session, base_url: str, user_headers: dict):
    r = session.post(f"{base_url}/users/setIsActive", headers=user_headers, json={"user_id": "u2", "is_active": False})
    assert r.status_code == 401
