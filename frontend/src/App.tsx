import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ShopProvider } from '@/context/ShopContext'
import { AuthProvider } from '@/context/AuthContext'
import { SessionProvider } from '@/context/SessionContext'
import { PersonaProvider } from '@/context/PersonaContext'
import { ProtectedRoute } from '@/components/auth/ProtectedRoute'
import { Layout } from '@/components/layout/Layout'
import { LandingPage } from '@/pages/LandingPage'
import { AssistantPage } from '@/pages/AssistantPage'
import { RecommendationsPage } from '@/pages/RecommendationsPage'
import { ProfilePage } from '@/pages/ProfilePage'
import { HistoryPage } from '@/pages/HistoryPage'
import { SavedPage } from '@/pages/SavedPage'
import { CartPage } from '@/pages/CartPage'
import { SignInPage } from '@/pages/SignInPage'
import { SignUpPage } from '@/pages/SignUpPage'
import { PersonaOnboardPage } from '@/pages/PersonaOnboardPage'
import { NotFoundPage } from '@/pages/NotFoundPage'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
      <SessionProvider>
      <ShopProvider>
      <PersonaProvider>
      <BrowserRouter>
        <Layout>
          <Routes>
            {/* Public: the landing page, the auth screens, and the partner
                onboarding portal. Everything else requires a signed-in account. */}
            <Route path="/" element={<LandingPage />} />
            <Route path="/signin" element={<SignInPage />} />
            <Route path="/signup" element={<SignUpPage />} />
            <Route path="/partner/onboard" element={<PersonaOnboardPage />} />

            <Route
              path="/assistant"
              element={
                <ProtectedRoute>
                  <AssistantPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/recommendations"
              element={
                <ProtectedRoute>
                  <RecommendationsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <ProfilePage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/history"
              element={
                <ProtectedRoute>
                  <HistoryPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/saved"
              element={
                <ProtectedRoute>
                  <SavedPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/cart"
              element={
                <ProtectedRoute>
                  <CartPage />
                </ProtectedRoute>
              }
            />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Layout>
      </BrowserRouter>
      </PersonaProvider>
      </ShopProvider>
      </SessionProvider>
      </AuthProvider>
    </QueryClientProvider>
  )
}

export default App
