import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ShopProvider } from '@/context/ShopContext'
import { Layout } from '@/components/layout/Layout'
import { LandingPage } from '@/pages/LandingPage'
import { AssistantPage } from '@/pages/AssistantPage'
import { RecommendationsPage } from '@/pages/RecommendationsPage'
import { HistoryPage } from '@/pages/HistoryPage'
import { SavedPage } from '@/pages/SavedPage'
import { CartPage } from '@/pages/CartPage'
import { NotFoundPage } from '@/pages/NotFoundPage'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ShopProvider>
      <BrowserRouter>
        <Layout>
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/assistant" element={<AssistantPage />} />
            <Route path="/recommendations" element={<RecommendationsPage />} />
            <Route path="/history" element={<HistoryPage />} />
            <Route path="/saved" element={<SavedPage />} />
            <Route path="/cart" element={<CartPage />} />
            <Route path="*" element={<NotFoundPage />} />
          </Routes>
        </Layout>
      </BrowserRouter>
      </ShopProvider>
    </QueryClientProvider>
  )
}

export default App
