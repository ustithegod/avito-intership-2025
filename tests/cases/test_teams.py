import uuid

import pytest
import requests


@pytest.mark.e2e
def test_add_team_success(session: requests.Session, base_url: str):
    team = f"team-{uuid.uuid4().hex[:8]}"
    r = session.post(
        f"{base_url}/team/add",
        json={
            "team_name": team,
            "members": [
                {"user_id": "u1", "username": "Alice", "is_active": True},
                {"user_id": "u2", "username": "Bob", "is_active": False},
            ],
        },
    )
    assert r.status_code == 201
    data = r.json()
    assert data["team"]["team_name"] == team
    assert len(data["team"]["members"]) == 2


@pytest.mark.e2e
@pytest.mark.negative
def test_add_team_duplicate(session: requests.Session, base_url: str):
    team = f"team-{uuid.uuid4().hex[:8]}"
    payload = {
        "team_name": team,
        "members": [{"user_id": "u1", "username": "A", "is_active": True}],
    }
    r1 = session.post(f"{base_url}/team/add", json=payload)
    assert r1.status_code == 201
    r2 = session.post(f"{base_url}/team/add", json=payload)
    assert r2.status_code == 400
    assert r2.json()["error"]["code"] == "TEAM_EXISTS"


@pytest.mark.e2e
@pytest.mark.negative
def test_add_team_validation(session: requests.Session, base_url: str):
    # team_name слишком длинный
    r = session.post(
        f"{base_url}/team/add",
        json={
            "team_name": "x" * 32,
            "members": [{"user_id": "u1", "username": "A", "is_active": True}],
        },
    )
    assert r.status_code == 400
    assert r.json()["error"]["code"] == "VALIDATION_ERROR"

    # username слишком длинный
    team = f"team-{uuid.uuid4().hex[:8]}"
    r = session.post(
        f"{base_url}/team/add",
        json={
            "team_name": team,
            "members": [{"user_id": "u1", "username": "Y" * 64, "is_active": True}],
        },
    )
    assert r.status_code == 400
    assert r.json()["error"]["code"] == "VALIDATION_ERROR"


@pytest.mark.e2e
def test_get_team_success_and_not_found(
    session: requests.Session, base_url: str, user_headers: dict
):
    team = f"team-{uuid.uuid4().hex[:8]}"
    session.post(
        f"{base_url}/team/add",
        json={
            "team_name": team,
            "members": [{"user_id": "u1", "username": "A", "is_active": True}],
        },
    )

    r1 = session.get(
        f"{base_url}/team/get", headers=user_headers, params={"team_name": team}
    )
    assert r1.status_code == 200
    assert r1.json()["team_name"] == team

    r2 = session.get(
        f"{base_url}/team/get", headers=user_headers, params={"team_name": "no-such"}
    )
    assert r2.status_code == 404


@pytest.mark.e2e
@pytest.mark.negative
def test_get_team_missing_param(
    session: requests.Session, base_url: str, user_headers: dict
):
    r = session.get(f"{base_url}/team/get", headers=user_headers)
    assert r.status_code == 400
