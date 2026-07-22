import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ArrowRight, Sparkles, Bot, Zap, ShieldCheck } from 'lucide-react'
import { Hero } from '@/components/home/Hero'
import { Button } from '@/components/ui/Button'
import { Categories } from '@/components/home/Categories' 
import { Occasions } from '@/components/home/Occasions'

export function LandingPage() {
  return (
    <div className="relative w-full bg-slate-50/50 overflow-hidden">
      {/* GLOBAL BACKGROUND AMBIENT GLOW SYSTEM */}
      <div className="fixed top-0 left-1/2 -translate-x-1/2 w-[1200px] h-[600px] bg-gradient-to-b from-rose-500/5 via-pink-500/5 to-transparent blur-[160px] pointer-events-none -z-10" />

      {/* HERO SECTION */}
      <Hero />

      {/* CATEGORIES SECTION */}
      <Categories />

      {/* OCCASIONS SECTION */}
      <Occasions />

      {/* NEXT-LEVEL HACKATHON CTA BANNER */}
      <section id="hackathon" className="relative mx-auto max-w-7xl px-5 py-24 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 32, filter: 'blur(10px)' }}
          whileInView={{ opacity: 1, y: 0, filter: 'blur(0px)' }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.7, type: 'spring', stiffness: 50 }}
          className="relative overflow-hidden rounded-[36px] bg-slate-950 px-8 py-20 text-center sm:px-16 border border-slate-800/80 shadow-[0_32px_80px_-20px_rgba(244,63,94,0.25)] group"
        >
          {/* 1. ATMOSPHERIC BACKDROP LIGHT ORBS */}
          <motion.div 
            animate={{ 
              scale: [1, 1.25, 1],
              opacity: [0.4, 0.7, 0.4] 
            }}
            transition={{ duration: 8, repeat: Infinity, ease: 'easeInOut' }}
            className="pointer-events-none absolute -top-32 -right-32 h-96 w-96 rounded-full bg-gradient-to-br from-rose-500 via-pink-600 to-amber-500 blur-[100px]" 
          />
          <motion.div 
            animate={{ 
              scale: [1.2, 1, 1.2],
              opacity: [0.3, 0.6, 0.3] 
            }}
            transition={{ duration: 10, repeat: Infinity, ease: 'easeInOut' }}
            className="pointer-events-none absolute -bottom-32 -left-32 h-96 w-96 rounded-full bg-gradient-to-tr from-rose-600 to-amber-400 blur-[110px]" 
          />

          {/* 2. GLASS SHIMMER SWEEP EFFECT */}
          <div className="absolute inset-0 w-[200%] h-full top-0 left-[-100%] bg-gradient-to-r from-transparent via-white/10 to-transparent -skew-x-12 translate-x-0 group-hover:translate-x-[200%] transition-transform duration-[1500ms] ease-out pointer-events-none z-10" />

          {/* 3. MESH BACKGROUND GRID OVERLAY */}
          <div className="absolute inset-0 bg-[radial-gradient(rgba(255,255,255,0.1)_1px,transparent_1px)] [background-size:24px_24px] opacity-25 pointer-events-none" />

          {/* 4. CONTENT WRAPPER */}
          <div className="relative z-20 flex flex-col items-center">
            
            {/* FLOATING BADGE */}
            <motion.div 
              whileHover={{ scale: 1.05 }}
              className="inline-flex items-center gap-2 rounded-full border border-rose-500/40 bg-rose-500/10 px-5 py-2 text-xs font-bold text-rose-300 backdrop-blur-xl shadow-lg shadow-rose-950/50"
            >
              <Sparkles className="h-4 w-4 text-rose-400 animate-pulse" />
              <span className="uppercase tracking-widest text-[11px] font-black">Built for WeForShe HackerRamp 2026</span>
            </motion.div>

            {/* HEADLINE */}
            <h2 className="mx-auto mt-8 max-w-2xl text-4xl font-extrabold tracking-tight text-white sm:text-5xl lg:text-6xl font-sans leading-[1.05]">
              Your closet just got an <br />
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-rose-400 via-pink-300 to-amber-300">
                AI Personal Stylist
              </span>
            </h2>

            {/* SUBTITLE */}
            <p className="mx-auto mt-6 max-w-lg text-base sm:text-lg text-slate-300 font-normal leading-relaxed">
              Start a natural conversation and receive hyper-personalized outfit edits matched to your budget, vibe, and occasion in seconds.
            </p>

            {/* CTA BUTTON WRAPPER */}
            <div className="mt-10 flex flex-col sm:flex-row items-center gap-4">
              <Link to="/assistant">
                <motion.div
                  whileHover={{ scale: 1.06 }}
                  whileTap={{ scale: 0.96 }}
                  transition={{ type: 'spring', stiffness: 300, damping: 20 }}
                >
                  <Button size="lg" variant="primary" className="relative group/btn overflow-hidden bg-gradient-to-r from-rose-500 via-pink-600 to-rose-500 text-white font-extrabold px-8 py-6 rounded-2xl shadow-[0_15px_30px_-8px_rgba(244,63,94,0.5)] border border-rose-400/40 text-sm uppercase tracking-wider flex items-center gap-2.5">
                    <Bot className="h-5 w-5 text-rose-200" />
                    <span>Try AI Assistant</span>
                    <ArrowRight className="h-4 w-4 group-hover/btn:translate-x-1.5 transition-transform" />
                  </Button>
                </motion.div>
              </Link>
            </div>

            {/* TRUST SPECS INLINE BADGES */}
            <div className="mt-12 flex flex-wrap justify-center items-center gap-6 text-slate-400 text-xs font-semibold">
              <div className="flex items-center gap-1.5">
                <Zap className="h-3.5 w-3.5 text-amber-400" />
                <span>Instant Lookbooks</span>
              </div>
              <span className="hidden sm:inline text-slate-700">•</span>
              <div className="flex items-center gap-1.5">
                <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
                <span>Myntra Catalog Sync</span>
              </div>
            </div>

          </div>
        </motion.div>
      </section>
    </div>
  )
}