import { useState, useRef, useEffect } from 'react'
import type { MouseEvent } from 'react'
import { motion, useInView, useMotionValue, useSpring, useTransform } from 'framer-motion'
import type { Variants } from 'framer-motion'
import { Sparkles, ArrowRight, ArrowLeft } from 'lucide-react'

const OCCASIONS = [
  { id: 'wedding', num: '01', title: 'Wedding', description: 'Elegant ethnic looks curated for your special day.', prompt: 'I need a wedding outfit under ₹3000', insight: 'Trending wedding styles in your budget', img: 'src/assets/images/occasions/img1.png' },
  { id: 'office', num: '02', title: 'Office', description: 'Professional styles constructed with effortless confidence.', prompt: 'Suggest a professional office outfit', insight: 'AI matched for a professional look', img: 'src/assets/images/occasions/img2.png' },
  { id: 'college', num: '03', title: 'College', description: 'Comfortable everyday fashion dialed into your aesthetic.', prompt: 'Recommend a casual college outfit', insight: 'Popular among students this week', img: 'src/assets/images/occasions/img3.png' },
  { id: 'party', num: '04', title: 'Party', description: 'Stand out from the room with bold premium evening looks.', prompt: 'Need a party outfit under ₹4000', insight: 'Curated for your evening style', img: 'src/assets/images/occasions/img4.png' },
  { id: 'travel', num: '05', title: 'Travel', description: 'Comfort meets first-class styling on every single journey.', prompt: 'Suggest comfortable travel outfits', insight: 'Comfortable outfits for long journeys', img: 'src/assets/images/occasions/img5.png' },
  { id: 'festival', num: '06', title: 'Festival', description: 'Celebrate cultural heritage traditions with modern elegance.', prompt: 'Recommend a festive ethnic look', insight: 'Traditional styles selected by AI', img: 'src/assets/images/occasions/img6.png' },
  { id: 'datenight', num: '07', title: 'Date Night', description: 'Timeless luxury silhouettes crafted for unforgettable evenings.', prompt: 'Suggest a stylish date night outfit', insight: 'Elegant combinations picked for you', img: 'src/assets/images/occasions/img7.png' },
  { id: 'gym', num: '08', title: 'Gym & Active', description: 'High-performance textiles engineered for everyday lifestyle comfort.', prompt: 'Show premium activewear', insight: 'Activewear optimized for comfort', img: 'src/assets/images/occasions/img8.png' },
]

export function Occasions() {
  const containerRef = useRef<HTMLDivElement>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const isInView = useInView(containerRef, { once: true, amount: 0.05 })
  const [isAutoScrolling, setIsAutoScrolling] = useState(true)

  const handleOccasionClick = (promptText: string) => {
    sessionStorage.setItem('prefilled_ai_prompt', promptText)
    window.location.href = '/assistant'
  }

  useEffect(() => {
    const scrollContainer = scrollRef.current
    if (!scrollContainer || !isAutoScrolling) return

    let animationFrameId: number
    const scrollSpeed = 0.6 

    const scroll = () => {
      if (!scrollContainer) return
      scrollContainer.scrollLeft += scrollSpeed
      
      if (scrollContainer.scrollLeft >= scrollContainer.scrollWidth - scrollContainer.clientWidth - 2) {
        scrollContainer.scrollLeft = 0
      }
      animationFrameId = requestAnimationFrame(scroll)
    }

    animationFrameId = requestAnimationFrame(scroll)
    return () => cancelAnimationFrame(animationFrameId)
  }, [isAutoScrolling])

  const handleManualScroll = (direction: 'left' | 'right') => {
    const scrollContainer = scrollRef.current
    if (!scrollContainer) return

    setIsAutoScrolling(false)

    const cardWidth = 340 
    const gap = 32
    const scrollAmount = direction === 'left' ? -(cardWidth + gap) : (cardWidth + gap)

    scrollContainer.scrollBy({
      left: scrollAmount,
      behavior: 'smooth'
    })

    setTimeout(() => {
      setIsAutoScrolling(true)
    }, 1500)
  }

  return (
    /* Added top padding pt-16 lg:pt-24 and removed bottom padding (pb-0) */
    <section className="relative w-full bg-stone-50 pt-16 lg:pt-24 pb-0 overflow-hidden flex flex-col items-center border-t border-slate-100/60">
      
      {/* Luxury Color Grading Ambient Atmosphere */}
      <div className="absolute top-0 right-0 w-[900px] h-[700px] bg-gradient-to-b from-rose-500/5 via-pink-600/5 to-transparent blur-[180px] pointer-events-none -z-10" />
      <div className="absolute bottom-0 left-[-10%] w-[600px] h-[600px] bg-gradient-to-tr from-amber-400/5 via-transparent to-transparent blur-[140px] pointer-events-none -z-10" />

      {/* Adjusted top margin mt-8 lg:mt-12 for proper content flow */}
      <div ref={containerRef} className="w-full max-w-[1440px] px-6 lg:px-12 grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-start relative z-10 mt-8 lg:mt-12">
        
        {/* --- LEFT SIDE: THE FIXED INTRO --- */}
        <div className="lg:col-span-4 lg:sticky lg:top-24 flex flex-col items-start text-left self-start">
          <motion.div 
            initial={{ opacity: 0, x: -15 }}
            animate={isInView ? { opacity: 1, x: 0 } : {}}
            transition={{ duration: 0.5 }}
            className="flex items-center gap-2 mb-4"
          >
            <span className="h-[1px] w-6 bg-gradient-to-r from-rose-500 to-amber-500" />
            <span className="text-[10px] font-black tracking-[0.35em] text-transparent bg-clip-text bg-gradient-to-r from-rose-500 to-pink-600 uppercase">
              Curated Matrix
            </span>
          </motion.div>
          
          <motion.h2 
            initial={{ opacity: 0, y: 15 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.5, delay: 0.05 }}
            className="text-4xl sm:text-5xl lg:text-[3.5rem] font-black tracking-tight text-slate-900 font-sans leading-[0.95]"
          >
            SHOP BY <br className="hidden lg:block"/>OCCASION
          </motion.h2>
          
          <motion.p 
            initial={{ opacity: 0, y: 15 }}
            animate={isInView ? { opacity: 1, y: 0 } : {}}
            transition={{ duration: 0.5, delay: 0.1 }}
            className="mt-6 text-sm lg:text-base text-slate-400 font-normal leading-relaxed tracking-wide max-w-sm"
          >
            Let our generative design models parse your vibe and assemble precise outfit architectures for every conceptual setting.
          </motion.p>

          <div className="mt-8 flex items-center gap-2.5 bg-rose-500/10 border border-rose-500/20 rounded-full px-4 py-1.5 text-[10px] font-bold text-rose-600 tracking-wider uppercase select-none">
            <Sparkles className="h-3 w-3 fill-rose-500 text-rose-500" />
            <span>AI Style Engine Active</span>
          </div>
        </div>

        {/* --- RIGHT SIDE: CANVAS & CONTROLS --- */}
        <div className="lg:col-span-8 w-full flex flex-col gap-6">
          <div 
            ref={scrollRef}
            onMouseEnter={() => setIsAutoScrolling(false)}
            onMouseLeave={() => setIsAutoScrolling(true)}
            onTouchStart={() => setIsAutoScrolling(false)}
            className="w-full overflow-x-auto pb-0 pt-2 scrollbar-none snap-x snap-mandatory touch-pan-x flex gap-8 will-change-transform"
          >
            {OCCASIONS.map((occasion, index) => (
              <div 
                key={occasion.id} 
                onClick={() => handleOccasionClick(occasion.prompt)}
                className="w-[85vw] sm:w-[45vw] lg:w-[340px] flex-shrink-0 snap-center"
              >
                <OccasionCard occasion={occasion} index={index} isInView={isInView} />
              </div>
            ))}
          </div>

          {/* Navigation Controls */}
          <div className="flex items-center justify-end gap-3 px-2">
            <button
              onClick={() => handleManualScroll('left')}
              className="p-3 rounded-full border border-slate-200 bg-white text-slate-600 hover:bg-gradient-to-r hover:from-rose-500 hover:to-pink-600 hover:text-white hover:border-transparent active:scale-95 transition-all duration-200 shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-rose-500"
              aria-label="Scroll left"
            >
              <ArrowLeft className="h-4 w-4" />
            </button>
            <button
              onClick={() => handleManualScroll('right')}
              className="p-3 rounded-full border border-slate-200 bg-white text-slate-600 hover:bg-gradient-to-r hover:from-rose-500 hover:to-pink-600 hover:text-white hover:border-transparent active:scale-95 transition-all duration-200 shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-rose-500"
              aria-label="Scroll right"
            >
              <ArrowRight className="h-4 w-4" />
            </button>
          </div>
        </div>

      </div>
    </section>
  )
}

/* --- THE PREMIUM ASYMMETRIC MYNTRA LOOKBOOK CARD --- */
function OccasionCard({ occasion, index, isInView }: { occasion: any, index: number, isInView: boolean }) {
  const cardRef = useRef<HTMLDivElement>(null)
  const [isHovered, setIsHovered] = useState(false)

  const x = useMotionValue(0)
  const y = useMotionValue(0)
  const mouseX = useMotionValue(0)
  const mouseY = useMotionValue(0)

  const rotateX = useSpring(useTransform(y, [-0.5, 0.5], [5, -5]), { stiffness: 150, damping: 24 })
  const rotateY = useSpring(useTransform(x, [-0.5, 0.5], [-5, 5]), { stiffness: 150, damping: 24 })
  const imgScale = useSpring(isHovered ? 1.10 : 1.02, { stiffness: 90, damping: 22 })

  const handleMouseMove = (e: MouseEvent<HTMLDivElement>) => {
    if (!cardRef.current) return
    const rect = cardRef.current.getBoundingClientRect()
    const mouseXRelative = e.clientX - rect.left
    const mouseYRelative = e.clientY - rect.top
    
    x.set(mouseXRelative / rect.width - 0.5)
    y.set(mouseYRelative / rect.height - 0.5)
    mouseX.set(mouseXRelative)
    mouseY.set(mouseYRelative)
  }

  const handleMouseLeave = () => {
    setIsHovered(false)
    x.set(0)
    y.set(0)
  }

  const revealVariants: Variants = {
    hidden: { opacity: 0, y: 50, filter: 'blur(10px)' },
    visible: { 
      opacity: 1, y: 0, filter: 'blur(0px)',
      transition: { type: 'spring', stiffness: 55, damping: 16, delay: (index % 4) * 0.05 }
    }
  }

  return (
    <motion.div
      ref={cardRef}
      variants={revealVariants}
      initial="hidden"
      animate={isInView ? "visible" : "hidden"}
      onMouseMove={handleMouseMove}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={handleMouseLeave}
      style={{ rotateX, rotateY, transformStyle: 'preserve-3d' }}
      whileHover={{ y: -6 }}
      transition={{ type: 'spring', stiffness: 220, damping: 26 }}
      className="relative aspect-[4/5.5] rounded-[28px] overflow-hidden bg-slate-900 shadow-sm cursor-pointer border border-slate-200/10 select-none group will-change-transform outline-none focus-visible:ring-2 focus-visible:ring-rose-500"
      tabIndex={0}
      aria-label={`Curate ${occasion.title} Outfits`}
    >
      <motion.img 
        src={occasion.img}
        alt={occasion.title}
        loading="lazy"
        style={{ scale: imgScale }}
        className="absolute inset-0 h-full w-full object-cover object-center pointer-events-none brightness-[0.85] group-hover:brightness-100 transition-all duration-700"
      />

      {/* Interactive Cursor Glow */}
      <motion.div 
        style={{
          background: useTransform(
            [mouseX, mouseY], 
            ([mx, my]) => `radial-gradient(250px circle at ${mx}px ${my}px, rgba(244,63,94,0.22), transparent 85%)`
          )
        }}
        className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none z-20"
      />

      <div className="absolute inset-0 bg-gradient-to-b from-black/10 via-black/5 to-black/85 pointer-events-none z-10" />

      <div className="absolute top-5 inset-x-5 z-20 flex justify-between items-center" style={{ transform: 'translateZ(20px)' }}>
        <span className="font-mono text-[10px] font-bold tracking-widest text-white/40 bg-black/30 backdrop-blur-md px-2 py-0.5 rounded">
          {occasion.num}
        </span>
        <div className="flex items-center gap-1 px-2.5 py-1 rounded-full border border-rose-500/30 bg-black/40 backdrop-blur-md text-[9px] font-black tracking-widest text-rose-400 uppercase shadow-sm">
          <Sparkles className="h-2.5 w-2.5 text-rose-400 fill-rose-400" />
          <span>AI Spec</span>
        </div>
      </div>

      <div className="absolute inset-x-0 bottom-0 p-6 z-20 flex flex-col justify-end overflow-hidden" style={{ transform: 'translateZ(30px)' }}>
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: isHovered ? 1 : 0, y: isHovered ? 0 : 10 }}
          transition={{ duration: 0.3, ease: 'easeOut' }}
          className="mb-2.5 text-[9px] font-black tracking-wider text-rose-300 uppercase bg-rose-950/70 backdrop-blur-md px-2.5 py-1.5 rounded-xl border border-rose-500/30 w-max"
        >
          ✨ {occasion.insight}
        </motion.div>

        <div className="flex flex-col gap-0.5">
          <h3 className="text-xl font-black text-white tracking-wide font-sans">
            {occasion.title}
          </h3>
          <p className="text-xs text-white/75 font-normal leading-relaxed tracking-wide opacity-90 group-hover:opacity-100 transition-opacity">
            {occasion.description}
          </p>
        </div>

        <motion.div 
          initial={{ opacity: 0, height: 0, marginTop: 0 }}
          animate={{ opacity: isHovered ? 1 : 0, height: isHovered ? 'auto' : 0, marginTop: isHovered ? 14 : 0 }}
          transition={{ duration: 0.3, ease: 'easeInOut' }}
          className="border-t border-white/10 pt-3 flex items-center justify-between group/btn"
        >
          <span className="text-[10px] font-mono tracking-wide text-white/50 truncate mr-2">
            "{occasion.prompt}"
          </span>
          <span className="p-1 rounded-full bg-white text-slate-900 group-hover/btn:bg-rose-500 group-hover/btn:text-white transition-colors duration-200 flex-shrink-0">
            <ArrowRight className="h-3 w-3" />
          </span>
        </motion.div>
      </div>

      <div className="absolute inset-0 rounded-[28px] opacity-0 group-hover:opacity-100 transition-opacity duration-300 -z-10 shadow-[0_30px_60px_-15px_rgba(244,63,94,0.25)]" />
    </motion.div>
  )
}