export const currentUser = {
  id: import.meta.env.VITE_CURRENT_USER_ID ?? '00000000-0000-0000-0000-000000000001',
  handle: import.meta.env.VITE_CURRENT_USER_HANDLE ?? '@taro_tanaka',
  displayName: import.meta.env.VITE_CURRENT_USER_DISPLAY_NAME ?? '田中 太郎',
  bio:
    import.meta.env.VITE_CURRENT_USER_BIO ?? 'プロダクト開発とユーザー体験について考えています。',
  avatar: import.meta.env.VITE_CURRENT_USER_AVATAR ?? '太',
}

export const showDemoUsers = !import.meta.env.PROD
