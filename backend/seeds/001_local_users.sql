INSERT INTO users (id, handle, display_name, bio)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'taro_tanaka', '田中 太郎', 'Webアプリケーションをつくっています。'),
    ('00000000-0000-0000-0000-000000000002', 'hanako_s', '鈴木 花子', 'デザインとプロダクトづくりが好きです。'),
    ('00000000-0000-0000-0000-000000000003', 'ken_yamamoto', '山本 健', 'バックエンドと開発体験を改善しています。'),
    ('00000000-0000-0000-0000-000000000004', 'sho_sato', '佐藤 翔', '小さく試して、学びながら開発しています。'),
    ('00000000-0000-0000-0000-000000000005', 'product_team', 'プロダクト開発部', 'チームでよりよいプロダクトをつくります。')
ON CONFLICT (id) DO UPDATE
SET
    handle = EXCLUDED.handle,
    display_name = EXCLUDED.display_name,
    bio = EXCLUDED.bio;
