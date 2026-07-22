import { useState, type FormEvent } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useAuth } from '@/context/AuthContext'

const inputClass =
  'w-full rounded-xl border border-line bg-white px-4 py-2.5 text-sm text-ink-900 outline-none transition-colors focus:border-brand-500'

export function SignInPage() {
  const { signIn } = useAuth()
  const location = useLocation()
  // ProtectedRoute records the page the shopper was trying to reach, so signing
  // in returns them there (e.g. the Assistant) instead of dropping them on Home.
  const from = (location.state as { from?: string } | null)?.from
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await signIn(email.trim(), password)

      // Back to the guarded page they wanted, otherwise Home. A full navigation
      // re-initialises the app under the new identity; the app-wide onboarding
      // gate then asks for personalization details only when this account's
      // profile is incomplete — so a complete profile lands straight on the
      // page, and an incomplete one gets onboarding first.
      window.location.assign(from && from !== '/signin' ? from : '/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not sign you in.')
      setSubmitting(false)
    }
  }

  return (
    <div className="mx-auto flex min-h-[calc(100vh-10rem)] max-w-md items-center px-5 py-10">
      <motion.div
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        className="w-full rounded-2xl border border-line bg-white p-7 shadow-sm sm:p-8"
      >
        <span className="flex h-11 w-11 items-center justify-center rounded-full bg-gradient-to-br from-brand-50 to-brand-100">
          <Sparkles className="h-5 w-5 text-brand-500" />
        </span>
        <h1 className="mt-4 font-display text-2xl font-bold text-ink-900">Welcome back</h1>
        <p className="mt-1 text-sm text-ink-500">Sign in to pick up where you left off.</p>

        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-ink-900">Email</span>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              autoComplete="email"
              required
              className={inputClass}
            />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-ink-900">Password</span>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Your password"
              autoComplete="current-password"
              required
              className={inputClass}
            />
          </label>

          {error && (
            <p className="rounded-lg bg-brand-50 px-3 py-2 text-sm text-brand-600">{error}</p>
          )}

          <Button type="submit" size="lg" className="w-full" disabled={submitting}>
            {submitting ? 'Signing in…' : 'Sign in'}
          </Button>
        </form>

        <p className="mt-5 text-center text-sm text-ink-500">
          New here?{' '}
          <Link to="/signup" className="font-semibold text-brand-500 hover:text-brand-600">
            Create an account
          </Link>
        </p>
      </motion.div>
    </div>
  )
}
