CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_teams_created_at ON teams(created_at);

CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_users_team FOREIGN KEY (team_name)
        REFERENCES teams(team_name)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_name, is_active) WHERE is_active = true;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP,
    CONSTRAINT fk_pr_author FOREIGN KEY (author_id)
        REFERENCES users(user_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_pr_status CHECK (status IN ('OPEN', 'MERGED'))
);

CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pr_created_at ON pull_requests(created_at);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL,
    reviewer_id VARCHAR(255) NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (pull_request_id, reviewer_id),
    CONSTRAINT fk_pr_reviewers_pr FOREIGN KEY (pull_request_id)
        REFERENCES pull_requests(pull_request_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_pr_reviewers_user FOREIGN KEY (reviewer_id)
        REFERENCES users(user_id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_assigned_at ON pr_reviewers(assigned_at);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_pull_requests_updated_at ON pull_requests;
CREATE TRIGGER update_pull_requests_updated_at BEFORE UPDATE ON pull_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE teams IS 'Stores team information';
COMMENT ON TABLE users IS 'Stores user information with team membership and active status';
COMMENT ON TABLE pull_requests IS 'Stores pull request information with status tracking';
COMMENT ON TABLE pr_reviewers IS 'Junction table linking pull requests to their assigned reviewers (max 2 per PR)';

COMMENT ON COLUMN users.is_active IS 'Only active users can be assigned as reviewers';
COMMENT ON COLUMN pull_requests.status IS 'PR status: OPEN or MERGED. MERGED PRs cannot have reviewers modified';
