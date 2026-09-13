ALTER TABLE public.posts
    ADD CONSTRAINT posts_author_id_fkey
    FOREIGN KEY (author_id) REFERENCES public.users (id)
    ON DELETE RESTRICT;

ALTER TABLE public.follows
    ADD CONSTRAINT follows_follower_id_fkey
    FOREIGN KEY (follower_id) REFERENCES public.users (id)
    ON DELETE CASCADE;

ALTER TABLE public.follows
    ADD CONSTRAINT follows_followee_id_fkey
    FOREIGN KEY (followee_id) REFERENCES public.users (id)
    ON DELETE CASCADE;
