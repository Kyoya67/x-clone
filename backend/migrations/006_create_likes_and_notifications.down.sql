DROP INDEX IF EXISTS public.notifications_recipient_created_at_idx;
DROP INDEX IF EXISTS public.notifications_like_unique_idx;
DROP INDEX IF EXISTS public.notifications_follow_unique_idx;
DROP TABLE IF EXISTS public.notifications;
DROP INDEX IF EXISTS public.post_likes_user_id_idx;
DROP TABLE IF EXISTS public.post_likes;
