CREATE TABLE IF NOT EXISTS users (
    uuid VARCHAR(36) NOT NULL UNIQUE,
    created_at INT(11) NOT NULL, -- UNIX time
    created_by_uuid VARCHAR(36) NOT NULL,
    updated_at INT(11), -- UNIX time
    updated_by_uuid VARCHAR(36),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    hashed_password VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    self_description TEXT(100),
    PRIMARY KEY(uuid)
);

CREATE TABLE IF NOT EXISTS user_user_follows (
    follower_uuid VARCHAR(36) NOT NULL,
    followed_uuid VARCHAR(36) NOT NULL,
    PRIMARY KEY(follower_uuid, followed_uuid),
    FOREIGN KEY(follower_uuid) REFERENCES srcabl_users.users(uuid),
    FOREIGN KEY(followed_uuid) REFERENCES srcabl_users.users(uuid)
);

CREATE TABLE IF NOT EXISTS user_source_follows (
    follower_uuid VARCHAR(36) NOT NULL,
    followed_uuid VARCHAR(36) NOT NULL,
    PRIMARY KEY(follower_uuid, followed_uuid),
    FOREIGN KEY(follower_uuid) REFERENCES srcabl_users.users(uuid),
    FOREIGN KEY(followed_uuid) REFERENCES srcabl_sources.sources(uuid)
);
