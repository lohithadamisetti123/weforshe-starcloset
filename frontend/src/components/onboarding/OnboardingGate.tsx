import { useState } from 'react'
import { useLocation } from 'react-router-dom'
import { AnimatePresence, motion } from 'framer-motion'
import { OnboardingFlow } from '@/components/onboarding/OnboardingFlow'
import { useProfile } from '@/hooks/useProfile'
import { useAuth } from '@/context/AuthContext'

// OnboardingGate shows the personalization questionnaire exactly once, to a
// signed-in shopper who has not completed it yet.
//
// Authentication is the precondition. A signed-out visitor is never asked:
// personalization belongs to an account, and a guest session is not one. This
// is enforced by *not rendering the inner component at all* when signed out,
// which also means no /profile/status request is made for anonymous visitors.
// (The status endpoint answers for guest sessions too — a guest is a real user
// row with no profile — so calling it while signed out is exactly what used to
// make the modal appear before sign-in.)
//
// Who sees it: a signed-in account whose profile has never been completed
// (`onboarded === false`) — i.e. straight after sign up, or after signing in to
// an account with an incomplete profile. A shopper with a complete profile
// never sees it again.
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
  const [dismissed, setDismissed] = useState(false)

  // onboarded === null while the status is still loading — render nothing so a
  // returning shopper never sees a flash of the modal.
  const open = onboarded === false && !dismissed

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-[100] overflow-y-auto bg-ink-900/40 backdrop-blur-sm"
        >
          <OnboardingFlow
            mode="onboarding"
            saving={saving}
            onSave={save}
            onDone={() => setDismissed(true)}
            onSkip={() => setDismissed(true)}
          />
        </motion.div>
      )}
    </AnimatePresence>
  )
}
