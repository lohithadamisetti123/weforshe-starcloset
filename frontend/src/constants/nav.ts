export interface NavLink {
  label: string
  to: string
}

export const navLinks: NavLink[] = [
  { label: 'Home', to: '/' },
  { label: 'Assistant', to: '/assistant' },
  { label: 'Recommendations', to: '/recommendations' },
  { label: 'Saved', to: '/saved' },
  { label: 'Cart', to: '/cart' },
]
