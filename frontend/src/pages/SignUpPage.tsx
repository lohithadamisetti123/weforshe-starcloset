import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { Sparkles } from 'lucide-react'
import { Button } from '@/components/ui/Button'
import { useAuth } from '@/context/AuthContext'

const inputClass =
  'w-full rounded-xl border border-line bg-white px-4 py-2.5 text-sm text-ink-900 outline-none transition-colors focus:border-brand-500'

export function SignUpPage() {
  const { signUp } = useAuth()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)

    if (name.trim().length < 2) return setError('Please enter your name.')
    if (password.length < 8) return setError('Password must be at least 8 characters.')

    setSubmitting(true)
    try {
      await signUp(name.trim(), email.trim(), password)
      // New account → onboarding. A full navigation re-initialises the app under
      // the new identity; the onboarding gate then appears because a brand-new
      // account has no profile yet.
      window.location.assign('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create your account.')
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
        <h1 className="mt-4 font-display text-2xl font-bold text-ink-900">Create your account</h1>
        <p className="mt-1 text-sm text-ink-500">Save your style and get recommendations made for you.</p>

        <form onSubmit={handleSubmit} className="mt-6 space-y-4">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-ink-900">Name</span>
            <input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Your name"
              autoComplete="name"
              className={inputClass}
            />
          </label>
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
              placeholder="At least 8 characters"
              autoComplete="new-password"
              required
              className={inputClass}
            />
          </label>

          {error && (
            <p className="rounded-lg bg-brand-50 px-3 py-2 text-sm text-brand-600">{error}</p>
          )}

          <Button type="submit" size="lg" className="w-full" disabled={submitting}>
            {submitting ? 'Creating account…' : 'Create account'}
          </Button>
        </form>

        <p className="mt-5 text-center text-sm text-ink-500">
          Already have an account?{' '}
          <Link to="/signin" className="font-semibold text-brand-500 hover:text-brand-600">
            Sign in
          </Link>
        </p>
      </motion.div>
    </div>
  )
}
