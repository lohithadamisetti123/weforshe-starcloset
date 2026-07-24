import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'
import { hasPartnerAccess } from '@/lib/authz'

// RequireRole is the reusable partner/admin route guard. A shopper (or guest,
// or signed-out visitor) who navigates directly to a partner URL is redirected
// to Home rather than shown the portal — the client mirror of the backend's
// RequireRoles guard. It never renders its children for an unauthorized user,
// so the onboarding UI is never mounted for a normal user even briefly.
export function RequireRole({ children }: { children: ReactNode }) {
  const { user } = useAuth()

  if (!hasPartnerAccess(user?.role)) {
    return <Navigate to="/" replace />
  }
  return <>{children}</>
}
