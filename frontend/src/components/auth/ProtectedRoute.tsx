import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'

// ProtectedRoute gates a route behind authentication. An unauthenticated shopper
// is sent to Sign In, with the page they wanted preserved in router state so the
// sign-in flow can send them back there afterwards.
//
// It is applied only to routes that genuinely represent "your account" (the
// Profile page), so the guest browsing experience — home, assistant,
// recommendations, cart, wishlist, collections — is entirely unchanged.
export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth()
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/signin" replace state={{ from: location.pathname }} />
  }
  return <>{children}</>
}
