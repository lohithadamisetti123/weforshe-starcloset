export interface NavLink {
  label: string
  to: string
}

export const navLinks: NavLink[] = [
  { label: 'Home', to: '/' },
  { label: 'Stylist Studio', to: '/assistant' },
  { label: 'Cart', to: '/cart' },
  { label: 'Profile', to: '/profile' },
  { label: 'Recommendations', to: '/recommendations' },
  { label: 'Saved', to: '/saved' },
]
