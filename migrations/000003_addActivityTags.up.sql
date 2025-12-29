CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE activity_tags (
    activity_id INTEGER REFERENCES activities(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (activity_id, tag_id)
);

-- Indexes for performance
CREATE INDEX idx_activities_user_date ON activities(user_id, activity_date);
CREATE INDEX idx_activity_tags_activity ON activity_tags(activity_id);
CREATE INDEX idx_activity_tags_tag ON activity_tags(tag_id);
CREATE INDEX idx_tags_name ON tags(name);