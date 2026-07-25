import type { ReactNode } from 'react'
import { motion } from 'framer-motion'
import { Navbar } from '@/components/layout/Navbar'
import { Footer } from '@/components/layout/Footer'
import { OnboardingGate } from '@/components/onboarding/OnboardingGate'

export function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-surface">
      <Navbar />
      <motion.main
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: 'easeOut' }}
        className="flex-1 pt-20"
      >
        {children}
      </motion.main>
      <Footer />
      {/* Mandatory one-time profile setup for freshly created accounts. Renders
          nothing for guests, returning shoppers, or while status loads. */}
      <OnboardingGate />
    </div>
  )
}
