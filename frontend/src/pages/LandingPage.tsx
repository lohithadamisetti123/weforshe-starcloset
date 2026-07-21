import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ArrowRight, Sparkles } from 'lucide-react'
import { Hero } from '@/components/home/Hero'
import { Button } from '@/components/ui/Button'
import { Categories } from '@/components/home/Categories' 
import { Occasions } from '@/components/home/Occasions'
export function LandingPage() {
  return (
    <div>
      <Hero />
      <Categories />
      <Occasions />

      <section id="hackathon" className="mx-auto max-w-7xl px-5 py-20 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.55 }}
          className="relative overflow-hidden rounded-xl bg-gradient-to-br from-ink-900 via-ink-900 to-brand-700 px-8 py-16 text-center sm:px-16"
        >
          <div className="pointer-events-none absolute -top-24 right-0 h-72 w-72 rounded-full bg-brand-500/30 blur-3xl" />
          <div className="pointer-events-none absolute -bottom-24 left-0 h-72 w-72 rounded-full bg-brand-500/20 blur-3xl" />
          <div className="relative">
            <div className="mx-auto inline-flex items-center gap-2 rounded-full border border-white/20 bg-white/10 px-4 py-1.5 text-xs font-semibold text-white">
              <Sparkles className="h-3.5 w-3.5" />
              Built for WeForShe HackerRamp 2026
            </div>
            <h2 className="mx-auto mt-6 max-w-xl text-3xl font-bold tracking-tight text-white sm:text-4xl">
              Your closet just got an AI stylist
            </h2>
            <p className="mx-auto mt-4 max-w-md text-white/70">
              Start a conversation and get a personalized outfit edit in seconds.
            </p>
            <Link to="/assistant" className="mt-8 inline-block">
              <Button size="lg" variant="primary">
                Try AI Assistant
                <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
          </div>
        </motion.div>
      </section>
    </div>
  )
}
