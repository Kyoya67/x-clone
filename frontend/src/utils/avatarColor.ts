const avatarClasses = [
  'avatar-blue',
  'avatar-pink',
  'avatar-green',
  'avatar-orange',
  'avatar-purple',
  'avatar-cyan',
]

export function avatarColorClass(seed: string) {
  const value = Array.from(seed).reduce((sum, char) => sum + char.charCodeAt(0), 0)
  return avatarClasses[value % avatarClasses.length]
}
