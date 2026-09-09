ALTER TABLE posts
    ADD CONSTRAINT posts_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES users (id)
    ON DELETE RESTRICT;

ALTER TABLE follows
    ADD CONSTRAINT follows_follower_id_fkey
    FOREIGN KEY (follower_id) REFERENCES users (id)
    ON DELETE CASCADE;

ALTER TABLE follows
    ADD CONSTRAINT follows_followee_id_fkey
    FOREIGN KEY (followee_id) REFERENCES users (id)
    ON DELETE CASCADE;
