import { Link } from 'react-router-dom'
import { Sparkles } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'

// Paths that require an account. While signed out these are dropped from the
// footer, for the same reason they're dropped from the navbar: they'd be dead
// ends straight into the sign-in screen. Anchor links ('/#about') and Home stay.
const PROTECTED_PATHS = ['/assistant', '/recommendations', '/saved', '/cart', '/profile', '/history']

const columns = [
  {
    title: 'Product',
    links: [
      { label: 'AI Assistant', to: '/assistant' },
      { label: 'Recommendations', to: '/recommendations' },
      { label: 'Saved', to: '/saved' },
      { label: 'Cart', to: '/cart' },
    ],
  },
  {
    title: 'Company',
    links: [
      { label: 'About', to: '/#about' },
      { label: 'Hackathon', to: '/#hackathon' },
    ],
  },
  {
    title: 'Legal',
    links: [
      { label: 'Privacy', to: '/#privacy' },
      { label: 'Terms', to: '/#terms' },
    ],
  },
]

export function Footer() {
  const { isAuthenticated } = useAuth()

  // Drop account-only links while signed out, then drop any column left empty
  // so the footer never renders a bare heading.
  const visibleColumns = columns
    .map((col) => ({
      ...col,
      links: isAuthenticated ? col.links : col.links.filter((l) => !PROTECTED_PATHS.includes(l.to)),
    }))
    .filter((col) => col.links.length > 0)

  return (
    <footer className="border-t border-line bg-white">
      <div className="mx-auto max-w-7xl px-5 py-14 lg:px-8">
        <div className="grid grid-cols-2 gap-10 md:grid-cols-5">
          <div className="col-span-2">
            <Link to="/" className="flex items-center gap-2">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-brand-500 to-brand-700">
                <Sparkles className="h-4 w-4 text-white" />
              </span>
              <span className="font-display text-base font-bold text-ink-900">
                Myntra <span className="text-brand-500">Bharat AI</span>
              </span>
            </Link>
            <p className="mt-4 max-w-xs text-sm leading-relaxed text-ink-500">
              Your personal AI fashion stylist — built for Myntra WeForShe HackerRamp 2026.
            </p>
          </div>
          {visibleColumns.map((col) => (
            <div key={col.title}>
              <p className="font-display text-sm font-semibold text-ink-900">{col.title}</p>
              <ul className="mt-4 space-y-3">
                {col.links.map((link) => (
                  <li key={link.label}>
                    <Link to={link.to} className="text-sm text-ink-500 transition-colors hover:text-brand-500">
                      {link.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        <div className="mt-12 flex flex-col items-center justify-between gap-4 border-t border-line pt-6 text-xs text-ink-500 sm:flex-row">
          <p>© 2026 Myntra Bharat AI. Built for WeForShe HackerRamp — a hackathon concept, not affiliated with production Myntra.</p>
          <p>Made with React, TypeScript &amp; Framer Motion</p>
        </div>
      </div>
    </footer>
  )
}
