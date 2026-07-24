import { useState, useRef } from 'react'
import { motion, useInView } from 'framer-motion'
import type { Variants } from 'framer-motion'
import { Sparkles, ArrowUpRight, Bot, Wand2 } from 'lucide-react'
import img1 from '@/assets/images/categories/img1.png'
import img2 from '@/assets/images/categories/img2.png'
import img3 from '@/assets/images/categories/img3.png'
import img4 from '@/assets/images/categories/img4.png'
import img6 from '@/assets/images/categories/img6.png'
import img7 from '@/assets/images/categories/img7.png'
import img8 from '@/assets/images/categories/img8.png'
const CATEGORIES = [
  {
    id: 'women',
    title: "Women's Fashion",
    tagline: 'Elegant styles curated for every moment.',
    img: img1,
    side: 'left',
  },
  {
    id: 'men',
    title: "Men's Fashion",
    tagline: 'Premium smart casual tailored to perfection.',
    img: img2,
    side: 'left',
  },
  {
    id: 'kids',
    title: "Kids' Collection",
    tagline: 'Happy premium design for little trendsetters.',
    img: img3,
    side: 'right',
  },
  {
    id: 'beauty',
    title: 'Beauty & Cosmetics',
    tagline: 'Luxury skincare and modern beauty formulas.',
    img: img4,
    side: 'right',
  },
  {
    id: 'footwear',
    title: 'Premium Footwear',
    tagline: 'Statement luxury sneakers and crafted shoes.',
    img: img6,
    side: 'left',
  },
  {
    id: 'watches',
    title: 'Luxury Watches',
    tagline: 'Timeless horology craftsmanship close-up.',
    img: img7,
    side: 'right',
  },
  {
    id: 'accessories',
    title: 'Elite Accessories',
    tagline: 'High-end statement designer handbags & sunglasses.',
    img: img8,
    side: 'right',
  },
  {
    id: 'ai-stylist',
    isAIPortal: true,
    side: 'right',
  },
]

export function Categories() {
  const containerRef = useRef(null)
  const isInView = useInView(containerRef, { once: true, amount: 0.05 })

  return (
    <section className="relative w-full bg-white pb-0 pt-0 overflow-hidden flex flex-col items-center">
      
      {/* --- LUXURY ATMOSPHERE BACKGROUND DECORATIONS --- */}
      <div className="absolute top-0 left-[-10%] w-[500px] h-[500px] bg-gradient-to-tr from-amber-200/20 via-rose-200/10 to-transparent blur-[140px] pointer-events-none rounded-full -z-10" />
      <div className="absolute bottom-0 right-[-10%] w-[600px] h-[600px] bg-gradient-to-bl from-amber-300/15 via-rose-200/10 to-transparent blur-[160px] pointer-events-none rounded-full -z-10" />
      
      <div className="absolute inset-0 pointer-events-none opacity-20 select-none bg-[radial-gradient(#d97706_1px,transparent_1px)] [background-size:40px_40px] [mask-image:radial-gradient(ellipse_at_center,black,transparent_75%)] -z-10" />

      {/* --- SECTION HEADER --- */}
      <div className="relative w-full max-w-[1440px] px-6 lg:px-12 mt-12 mb-12 lg:mb-16 text-center z-10 flex flex-col items-center">
        <motion.span 
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="text-xs font-black tracking-[0.25em] text-transparent bg-clip-text bg-gradient-to-r from-amber-600 via-rose-500 to-amber-700 uppercase mb-3"
        >
          Curated Exclusives
        </motion.span>
        
        <motion.h2 
          initial={{ opacity: 0, y: 15 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.05 }}
          className="text-4xl sm:text-5xl lg:text-6xl font-extrabold tracking-tight text-slate-900 font-sans leading-[1.1]"
        >
          SHOP BY CATEGORY
        </motion.h2>
        
        <motion.p 
          initial={{ opacity: 0, y: 15 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.1 }}
          className="mt-4 max-w-xl text-base sm:text-lg text-slate-500 font-normal leading-relaxed tracking-wide"
        >
          Explore fashion across every category with beautifully curated collections powered by AI.
        </motion.p>
      </div>

      {/* --- HIGH LEVEL SCROLL CONVERGENCE CONTAINER --- */}
      <div 
        ref={containerRef}
        className="relative w-full max-w-[1440px] px-0 sm:px-6 lg:px-12 z-10"
      >
        {/* Mobile View */}
        <div className="flex sm:hidden overflow-x-auto gap-5 px-6 pb-0 scrollbar-none snap-x snap-mandatory touch-pan-x">
          {CATEGORIES.map((category, index) => (
            <div key={category.id} className="w-[82vw] flex-shrink-0 snap-center">
              {category.isAIPortal ? (
                <AIPortalTrigger index={index} isInView={true} />
              ) : (
                <CategoryCard category={category} index={index} isInView={true} />
              )}
            </div>
          ))}
        </div>

        {/* Desktop Grid View */}
        <div className="hidden sm:grid grid-cols-2 lg:grid-cols-4 gap-8 overflow-visible">
          {CATEGORIES.map((category, index) => (
            category.isAIPortal ? (
              <AIPortalTrigger key={category.id} index={index} isInView={isInView} />
            ) : (
              <CategoryCard 
                key={category.id} 
                category={category} 
                index={index} 
                isInView={isInView} 
              />
            )
          ))}
        </div>
      </div>
    </section>
  )
}

/* --- ELEGANT BISCUIT & GOLD AI PORTAL TRIGGER --- */
function AIPortalTrigger({ index, isInView }: { index: number, isInView: boolean }) {
  const [isHovered, setIsHovered] = useState(false)

  const handleClick = () => {
    window.location.href = '/assistant'
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95, filter: 'blur(10px)' }}
      animate={isInView ? { opacity: 1, scale: 1, filter: 'blur(0px)' } : {}}
      transition={{ duration: 0.8, delay: (index % 4) * 0.08, type: 'spring' }}
      onClick={handleClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      whileHover={{ scale: 1.03, y: -4 }}
      className="relative aspect-[3/4] sm:aspect-[4/5] lg:aspect-[3/4.2] w-full h-full rounded-[28px] flex flex-col items-center justify-center cursor-pointer group select-none overflow-hidden bg-[#FBF8F3] border border-amber-200/60 shadow-sm transition-all duration-300"
      tabIndex={0}
      role="button"
      aria-label="Launch AI Stylist Assistant"
    >
      {/* 1. SOFT BISCUIT / GOLD ORBITAL LIGHT FIELD */}
      <motion.div 
        animate={{ 
          scale: isHovered ? [1.05, 1.25, 1.05] : [1, 1.1, 1],
          rotate: [0, 180, 360],
          opacity: isHovered ? 0.85 : 0.55
        }}
        transition={{ duration: 12, repeat: Infinity, ease: 'linear' }}
        className="absolute w-72 h-72 rounded-full bg-gradient-to-tr from-amber-200/80 via-rose-200/50 to-amber-300/70 blur-[50px] pointer-events-none"
      />

      {/* 2. REVOLVING FINE METALLIC RINGS */}
      <motion.div 
        animate={{ rotate: [0, 360] }}
        transition={{ duration: 16, repeat: Infinity, ease: 'linear' }}
        className="absolute w-60 h-60 rounded-full border border-dashed border-amber-400/30 pointer-events-none"
      />
      
      <motion.div 
        animate={{ rotate: [360, 0] }}
        transition={{ duration: 10, repeat: Infinity, ease: 'linear' }}
        className="absolute w-48 h-48 rounded-full border border-amber-300/40 border-t-rose-400/50 pointer-events-none"
      />

      {/* 3. CENTER TRIGGER NODE */}
      <div className="relative z-20 flex flex-col items-center gap-6">
        <motion.div
          animate={{ 
            scale: isHovered ? 1.12 : 1,
            y: isHovered ? -4 : 0
          }}
          transition={{ type: 'spring', stiffness: 300, damping: 18 }}
          className="relative w-20 h-20 sm:w-22 sm:h-22 rounded-full bg-stone-900/90 border border-amber-300/50 backdrop-blur-md shadow-[0_12px_30px_-6px_rgba(217,119,6,0.25)] flex items-center justify-center text-white overflow-hidden group-hover:border-amber-200 transition-colors"
        >
          {/* Subtle Inner Sweep */}
          <div className="absolute inset-0 bg-gradient-to-tr from-amber-400/20 via-transparent to-rose-400/20 opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
          
          {/* Animated Main Arrow */}
          <motion.div
            animate={{ 
              x: isHovered ? [0, 3, 0] : 0, 
              y: isHovered ? [0, -3, 0] : 0 
            }}
            transition={{ duration: 1.5, repeat: Infinity }}
            className="relative z-10 flex items-center justify-center"
          >
            <ArrowUpRight size={32} className="text-amber-200 group-hover:text-white transition-colors" />
          </motion.div>

          {/* Floating Sparkles */}
          <motion.div 
            animate={{ rotate: [0, 360] }}
            transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
            className="absolute inset-0 p-2 flex items-center justify-between pointer-events-none"
          >
            <Sparkles size={11} className="text-amber-300 opacity-90" />
            <Wand2 size={11} className="text-rose-300 opacity-90" />
          </motion.div>
        </motion.div>

        {/* 4. TYPOGRAPHY & INTENT */}
        <div className="text-center space-y-1.5 px-4 z-20">
          <motion.div 
            animate={{ y: isHovered ? -2 : 0 }}
            className="inline-flex items-center gap-1.5 text-[11px] font-black tracking-[0.2em] text-transparent bg-clip-text bg-gradient-to-r from-amber-700 via-rose-600 to-amber-800 uppercase"
          >
            <Bot size={13} className="text-amber-700" /> StarCloset Stylists
          </motion.div>
          <h3 className="text-xl sm:text-2xl font-extrabold text-stone-900 group-hover:text-amber-800 transition-colors font-sans">
            Meet Your Stylists
          </h3>
          <p className="text-xs text-stone-500 max-w-[190px] font-medium leading-relaxed mx-auto">
            Interactive lookbook generator tuned to your style
          </p>
        </div>
      </div>

      {/* 5. LIGHT SHIMMER SWEEP OVERLAY */}
      <div className="absolute inset-0 w-[200%] h-full top-0 left-[-100%] bg-gradient-to-r from-transparent via-amber-100/30 to-transparent -skew-x-12 translate-x-0 group-hover:translate-x-[200%] transition-transform duration-[1200ms] ease-out pointer-events-none z-20" />

      {/* 6. HOVER SHADOW EXPANSION */}
      <div className="absolute inset-0 rounded-[28px] ring-1 ring-amber-300/30 opacity-0 group-hover:opacity-100 transition-opacity duration-300 -z-10 shadow-[0_20px_45px_-12px_rgba(217,119,6,0.2)]" />
    </motion.div>
  )
}

/* --- CONVERGENT EDITORIAL CARD --- */
function CategoryCard({ category, index, isInView }: { category: any, index: number, isInView: boolean }) {
  const [isHovered, setIsHovered] = useState(false)

  const xOffset = category.side === 'left' ? -120 : 120

  const cardVariants: Variants = {
    hidden: { 
      opacity: 0, 
      x: xOffset,
      y: 40,
      filter: 'blur(16px)',
      scale: 0.95
    },
    visible: { 
      opacity: 1, 
      x: 0,
      y: 0,
      filter: 'blur(0px)',
      scale: 1,
      transition: { 
        type: 'spring', 
        stiffness: 45, 
        damping: 14, 
        delay: (index % 4) * 0.08
      }
    }
  }

  return (
    <motion.div
      initial="hidden"
      animate={isInView ? "visible" : "hidden"}
      variants={cardVariants}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      whileHover={{ scale: 1.03, y: -4 }}
      transition={{ type: 'spring', stiffness: 200, damping: 22 }}
      className="relative aspect-[3/4] sm:aspect-[4/5] lg:aspect-[3/4.2] rounded-[28px] overflow-hidden bg-slate-50 shadow-sm cursor-pointer border border-slate-200/40 select-none group will-change-transform outline-none focus-visible:ring-2 focus-visible:ring-rose-500"
      tabIndex={0}
      aria-label={`Browse ${category.title} Category`}
    >
      {/* 1. Large Dynamic Campaign Background Image */}
      <motion.img 
        src={category.img}
        alt={category.title}
        loading="lazy"
        animate={{ scale: isHovered ? 1.08 : 1 }}
        transition={{ duration: 0.8, ease: [0.25, 1, 0.5, 1] }}
        className="absolute inset-0 h-full w-full object-cover object-center pointer-events-none"
      />

      {/* 2. Custom Gradient Boundary Frame Overlay */}
      <div className="absolute inset-0 border border-transparent bg-gradient-to-tr from-white/10 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700 pointer-events-none rounded-[28px] z-20" />

      {/* 3. Luxury Cinematic Vignette Overlay */}
      <div className="absolute inset-0 bg-gradient-to-b from-black/5 via-transparent to-black/80 pointer-events-none z-10" />

      {/* 4. Glass Reflection Shimmer Sweep Effect */}
      <div className="absolute inset-0 w-[200%] h-full top-0 left-[-100%] bg-gradient-to-r from-transparent via-white/10 to-transparent -skew-x-12 translate-x-0 group-hover:translate-x-[200%] transition-transform duration-[1200ms] ease-out pointer-events-none z-20" />

      {/* 5. Editorial Typography Base Box */}
      <div className="absolute inset-x-0 bottom-0 p-6 sm:p-8 z-20 flex flex-col justify-end overflow-hidden">
        <motion.div
          animate={{ y: isHovered ? -2 : 0 }}
          transition={{ duration: 0.4, ease: [0.25, 1, 0.5, 1] }}
          className="flex flex-col gap-1.5"
        >
          {/* Category Title */}
          <h3 className="text-xl sm:text-2xl font-extrabold text-white tracking-wide">
            {category.title}
          </h3>

          {/* Tagline reveals gracefully on hover */}
          <motion.p 
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: isHovered ? 1 : 0, height: isHovered ? 'auto' : 0 }}
            transition={{ duration: 0.3, ease: 'easeInOut' }}
            className="text-xs sm:text-sm text-white/85 font-normal leading-relaxed tracking-wide"
          >
            {category.tagline}
          </motion.p>
        </motion.div>
      </div>

      {/* 6. Deep High-Tier Shadow Expansion Background Layer */}
      <div className="absolute inset-0 rounded-[28px] ring-1 ring-black/5 opacity-0 group-hover:opacity-100 transition-opacity duration-300 -z-10 shadow-[0_30px_60px_-15px_rgba(0,0,0,0.45)]" />
    </motion.div>
  )
}