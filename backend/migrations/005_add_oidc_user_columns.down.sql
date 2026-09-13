ALTER TABLE public.users
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS oidc_subject;

