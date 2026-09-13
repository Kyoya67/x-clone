CREATE TABLE IF NOT EXISTS public.post_likes (
    post_id UUID NOT NULL REFERENCES public.posts (id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES public.users (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX IF NOT EXISTS post_likes_user_id_idx ON public.post_likes (user_id);

CREATE TABLE IF NOT EXISTS public.notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_id UUID NOT NULL REFERENCES public.users (id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES public.users (id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('follow', 'like')),
    post_id UUID REFERENCES public.posts (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS notifications_follow_unique_idx
    ON public.notifications (recipient_id, actor_id, type)
    WHERE type = 'follow';

CREATE UNIQUE INDEX IF NOT EXISTS notifications_like_unique_idx
    ON public.notifications (recipient_id, actor_id, type, post_id)
    WHERE type = 'like';

CREATE INDEX IF NOT EXISTS notifications_recipient_created_at_idx
    ON public.notifications (recipient_id, created_at DESC);
