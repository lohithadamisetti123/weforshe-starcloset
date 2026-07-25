import { useLocation, useNavigate } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { ProfileSetup } from '@/components/onboarding/ProfileSetup'
import { useProfile } from '@/hooks/useProfile'
import { useAuth } from '@/context/AuthContext'

// OnboardingGate shows the mandatory profile setup exactly once, to a signed-in
// shopper who has not completed it yet.
//
// Authentication is the precondition. A signed-out visitor is never asked:
// personalization belongs to an account, and a guest session is not one. This
// is enforced by *not rendering the inner component at all* when signed out,
// which also means no /profile/status request is made for anonymous visitors.
//
// Who sees it: a signed-in account whose profile has never been completed
// (`onboarded === false`) — i.e. straight after sign up, or after signing in to
// an account with an incomplete profile. It is a full-screen takeover with no
// skip: the shopper completes setup once, is redirected Home, and never sees it
// again. While the status is loading (`onboarded === null`) nothing renders, so
// a returning shopper never sees a flash of the setup screen.
export function OnboardingGate() {
  const { isAuthenticated } = useAuth()
  const { pathname } = useLocation()

  // Never overlay the auth screens themselves.
  if (!isAuthenticated || pathname === '/signin' || pathname === '/signup') {
    return null
  }
  return <AuthedOnboardingGate />
}

// Rendered only for signed-in shoppers, so the profile lookup it performs never
// runs for an anonymous visitor.
function AuthedOnboardingGate() {
  const { onboarded, saving, save } = useProfile()
  const { user } = useAuth()
  const navigate = useNavigate()

  const open = onboarded === false

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.35 }}
          className="fixed inset-0 z-[100] overflow-y-auto"
          style={{ background: 'radial-gradient(120% 110% at 50% 0%, #FDFBF6 0%, #FAF2E2 60%, #F2E4C8 100%)' }}
        >
          <ProfileSetup
            userName={user?.name}
            saving={saving}
            onSave={save}
            onComplete={() => navigate('/')}
          />
        </motion.div>
      )}
    </AnimatePresence>
  )
}
