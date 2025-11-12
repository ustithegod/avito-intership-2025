CREATE INDEX idx_users_is_active ON users(is_active);

CREATE INDEX idx_teams_name ON teams(name);

CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);
