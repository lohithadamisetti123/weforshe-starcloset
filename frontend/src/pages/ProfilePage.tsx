import { useEffect, useState } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { MapPin, Pencil, ShoppingBag, Sparkles, User, Wallet } from 'lucide-react'
import { apiFetch } from '@/lib/api'
import { formatINR } from '@/lib/utils'
import { useProfile } from '@/hooks/useProfile'
import { OnboardingFlow } from '@/components/onboarding/OnboardingFlow'
import type { PastOrder, UserProfile } from '@/types'

// A read-only chip row for one preference group.
function ChipRow({ label, values }: { label: string; values: string[] }) {
  return (
    <div>
      <p className="text-xs font-semibold uppercase tracking-wide text-ink-500">{label}</p>
      {values.length === 0 ? (
        <p className="mt-1.5 text-sm text-ink-400">Not set</p>
      ) : (
        <div className="mt-2 flex flex-wrap gap-2">
          {values.map((v) => (
            <span
              key={v}
              className="rounded-full border border-line bg-surface px-3 py-1 text-sm font-medium text-ink-700"
            >
              {v}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

export function ProfilePage() {
  const { profile: hookProfile, loadProfile, saving, save } = useProfile()
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [orders, setOrders] = useState<PastOrder[]>([])
  const [editing, setEditing] = useState(false)
  const [loading, setLoading] = useState(true)

  // Load the full profile + purchase history on mount.
  useEffect(() => {
    let alive = true
    Promise.all([
      loadProfile(),
      apiFetch<PastOrder[]>('/profile/orders').catch(() => [] as PastOrder[]),
    ])
      .then(([p, o]) => {
        if (!alive) return
        setProfile(p)
        setOrders(o)
      })
      .finally(() => {
        if (alive) setLoading(false)
      })
    return () => {
      alive = false
    }
  }, [loadProfile])

  // Keep the local copy in sync if the hook refreshes it (e.g. after save).
  useEffect(() => {
    if (hookProfile) setProfile(hookProfile)
  }, [hookProfile])

  const budgetLabel =
    profile?.budgetMin != null && profile?.budgetMax != null
      ? `${formatINR(profile.budgetMin)} – ${formatINR(profile.budgetMax)}`
      : 'Not set'

  const onboarded = profile?.onboarded ?? false

  return (
    <div className="mx-auto max-w-5xl px-5 py-10 lg:px-8">
      <div className="flex flex-col gap-2">
        <p className="text-xs font-semibold uppercase tracking-widest text-brand-500">Your profile</p>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-ink-900 sm:text-4xl">
              {profile?.name ? profile.name : 'Style preferences'}
            </h1>
            <p className="mt-1 max-w-xl text-ink-500">
              Your assistant and recommendations use these automatically — no need to repeat yourself.
            </p>
          </div>
          {onboarded && (
            <button
              onClick={() => setEditing(true)}
              className="flex shrink-0 items-center gap-1.5 rounded-full bg-brand-500 px-4 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-600"
            >
              <Pencil className="h-4 w-4" />
              Edit Profile
            </button>
          )}
        </div>
      </div>

      {loading ? (
        <div className="mt-10 h-40 animate-pulse rounded-2xl border border-line bg-white" />
      ) : !onboarded ? (
        <div className="mt-10 flex flex-col items-center justify-center rounded-2xl border border-dashed border-line py-16 text-center">
          <span className="flex h-14 w-14 items-center justify-center rounded-full bg-gradient-to-br from-brand-50 to-brand-100">
            <Sparkles className="h-6 w-6 text-brand-500" />
          </span>
          <p className="mt-4 max-w-sm text-sm text-ink-500">
            You haven't set up your preferences yet. Tell us your style and we'll personalize everything.
          </p>
          <button
            onClick={() => setEditing(true)}
            className="mt-5 flex items-center gap-1.5 rounded-full bg-brand-500 px-5 py-2.5 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-brand-600"
          >
            <Sparkles className="h-4 w-4" />
            Set up preferences
          </button>
        </div>
      ) : (
        <>
          {/* Basics */}
          <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
            <div className="flex items-center gap-3 rounded-2xl border border-line bg-white p-4">
              <span className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-50 text-brand-500">
                <User className="h-5 w-5" />
              </span>
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-ink-500">Shopping for</p>
                <p className="text-sm font-semibold text-ink-900">{profile?.gender || 'Not set'}</p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-2xl border border-line bg-white p-4">
              <span className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-50 text-brand-500">
                <MapPin className="h-5 w-5" />
              </span>
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-ink-500">City</p>
                <p className="text-sm font-semibold text-ink-900">{profile?.city || 'Not set'}</p>
              </div>
            </div>
            <div className="flex items-center gap-3 rounded-2xl border border-line bg-white p-4">
              <span className="flex h-10 w-10 items-center justify-center rounded-full bg-brand-50 text-brand-500">
                <Wallet className="h-5 w-5" />
              </span>
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-ink-500">Budget</p>
                <p className="text-sm font-semibold text-ink-900">{budgetLabel}</p>
              </div>
            </div>
          </div>

          {/* Preference groups */}
          <div className="mt-4 grid grid-cols-1 gap-6 rounded-2xl border border-line bg-white p-6 sm:grid-cols-2">
            <ChipRow label="Styles" values={profile?.favoriteStyles ?? []} />
            <ChipRow label="Colours" values={profile?.favoriteColors ?? []} />
            <ChipRow label="Categories" values={profile?.favoriteCategories ?? []} />
            <ChipRow label="Occasions" values={profile?.favoriteOccasions ?? []} />
            <ChipRow label="Brands" values={profile?.favoriteBrands ?? []} />
            <ChipRow label="Sizes" values={profile?.sizes ?? []} />
          </div>

          {/* Purchase history */}
          <div className="mt-10 flex items-center gap-2">
            <ShoppingBag className="h-4 w-4 text-brand-500" />
            <p className="font-display text-lg font-semibold text-ink-900">Purchase history</p>
          </div>
          {orders.length === 0 ? (
            <p className="mt-3 text-sm text-ink-400">No orders yet.</p>
          ) : (
            <div className="mt-4 grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
              {orders.map((o) => (
                <div key={o.id} className="overflow-hidden rounded-xl border border-line bg-white">
                  <div className="aspect-[3/4] w-full overflow-hidden bg-surface">
                    {o.image ? (
                      <img src={o.image} alt={o.title} className="h-full w-full object-cover" loading="lazy" />
                    ) : (
                      <div className="flex h-full w-full items-center justify-center text-ink-300">
                        <ShoppingBag className="h-8 w-8" />
                      </div>
                    )}
                  </div>
                  <div className="p-3">
                    <p className="truncate text-xs font-semibold text-ink-500">{o.brand}</p>
                    <p className="line-clamp-2 text-sm font-medium text-ink-900">{o.title}</p>
                    <p className="mt-1 text-sm font-bold text-ink-900">{formatINR(o.price)}</p>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      {/* Edit / setup modal */}
      <AnimatePresence>
        {editing && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            className="fixed inset-0 z-[100] overflow-y-auto bg-ink-900/40 backdrop-blur-sm"
          >
            <OnboardingFlow
              mode="edit"
              initial={profile}
              saving={saving}
              onSave={async (req) => {
                const saved = await save(req)
                if (saved) setProfile(saved)
                return saved
              }}
              onDone={() => setEditing(false)}
              onClose={() => setEditing(false)}
            />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
