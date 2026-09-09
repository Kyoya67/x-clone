ALTER TABLE follows DROP CONSTRAINT IF EXISTS follows_followee_id_fkey;
ALTER TABLE follows DROP CONSTRAINT IF EXISTS follows_follower_id_fkey;
ALTER TABLE posts DROP CONSTRAINT IF EXISTS posts_author_id_fkey;
