import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Sparkles, Ticket, Percent, Zap } from 'lucide-react'

const BANNER_IMAGES = [
  'src/assets/images/hero/image1.png', 
  'src/assets/images/hero/image2.png',
  'src/assets/images/hero/image3.png',
  'src/assets/images/hero/image5.png',
]

const TICKER_ITEMS = [
  { text: "FLAT ₹400 OFF on First Order", icon: Ticket, code: "WELCOME400", highlight: "NEW USER EXCLUSIVE" },
  { text: "End of Season Sale: Up to 70% OFF", icon: Percent, code: "EOSS70", highlight: "LIMITED TIME ONLY" },
  { text: "Insane Lightning Deals Active Now", icon: Zap, code: "FLASHDEAL", highlight: "SELLING FAST" },
  { text: "Free Shipping on Orders Above ₹799", icon: Sparkles, code: "FREESHIP", highlight: "AUTOMATIC VALUE" },
]

export function Hero() {
  const [currentIndex, setCurrentIndex] = useState(0)

  useEffect(() => {
    const timer = setInterval(() => {
      setCurrentIndex((prevIndex) => (prevIndex + 1) % BANNER_IMAGES.length)
    }, 4000)

    return () => clearInterval(timer)
  }, [])

  const duplicatedTicker = [...TICKER_ITEMS, ...TICKER_ITEMS, ...TICKER_ITEMS]

  return (
    <section className="w-full bg-white overflow-hidden relative flex flex-col">
      {/* Background Subtle Gradient Glow */}
      <div className="absolute top-1/4 left-1/2 -translate-x-1/2 w-[90%] h-[400px] bg-gradient-to-r from-rose-400/10 via-amber-400/5 to-pink-500/10 blur-[120px] pointer-events-none -z-10" />

      {/* --- BANNER CONTAINER --- */}
      <div className="relative w-full mx-auto max-w-[2560px] aspect-[2560/1080] bg-white select-none">
        
        {/* Sliding Images Container */}
        <AnimatePresence mode="popLayout">
          <motion.img
            key={currentIndex}
            src={BANNER_IMAGES[currentIndex]}
            alt={`Hero Banner ${currentIndex + 1}`}
            initial={{ opacity: 0, x: '100%' }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: '-100%' }}
            transition={{ type: 'spring', stiffness: 180, damping: 26, opacity: { duration: 0.4 } }}
            className="absolute inset-0 h-full w-full object-cover object-top"
          />
        </AnimatePresence>

        {/* Dynamic Shadow Gradients */}
        <div className="absolute inset-0 bg-gradient-to-b from-black/5 via-transparent to-black/10 pointer-events-none" />

        {/* Large Minimalist Navigation Dots */}
        <div className="absolute bottom-6 left-1/2 flex -translate-x-1/2 gap-3 z-10">
          {BANNER_IMAGES.map((_, index) => (
            <button
              key={index}
              onClick={() => setCurrentIndex(index)}
              className={`h-2 rounded-full transition-all duration-300 ${
                index === currentIndex 
                  ? 'w-10 bg-white shadow-md shadow-black/10' 
                  : 'w-2 bg-white/50 hover:bg-white/80'
              }`}
              aria-label={`Go to slide ${index + 1}`}
            />
          ))}
        </div>
      </div>

      {/* --- DUAL-LAYERED HIGH LEVEL TICKER STRIP --- */}
      <div className="w-full bg-white flex flex-col relative">
        
        {/* Top Header Label Accent: Bridges gaps perfectly and locks onto the banner */}
        <div className="w-full bg-gradient-to-r from-rose-500 via-pink-500 to-amber-500 px-6 py-1.5 flex justify-between items-center text-[10px] sm:text-xs font-black tracking-widest text-white uppercase shadow-sm">
          <span>⚡ LIVE OFFERS & DEALS ⚡</span>
          <span className="opacity-90">CLICK COUPON TO COPY</span>
          <span>⚡ TIME-LIMITED STYLING ⚡</span>
        </div>

        {/* The Main High-Velocity Carousel: Doubled in vertical padding (`py-8`) */}
        <div className="w-full bg-white border-b border-rose-100/70 py-8 overflow-hidden relative flex items-center shadow-md">
          {/* Enhanced Glass Edge Overlays */}
          <div className="absolute left-0 top-0 bottom-0 w-32 bg-gradient-to-r from-white via-white/80 to-transparent z-10 pointer-events-none" />
          <div className="absolute right-0 top-0 bottom-0 w-32 bg-gradient-to-l from-white via-white/80 to-transparent z-10 pointer-events-none" />

          <motion.div 
            className="flex whitespace-nowrap gap-20 items-center px-6"
            animate={{ x: [0, -2200] }}
            transition={{
              ease: "linear",
              duration: 28,
              repeat: Infinity,
            }}
          >
            {duplicatedTicker.map((item, idx) => {
              const Icon = item.icon
              return (
                <div key={idx} className="flex items-center gap-6 flex-shrink-0 group cursor-pointer">
                  {/* High-Tier Oversized Icon Badge */}
                  <div className="p-3.5 rounded-2xl bg-gradient-to-tr from-rose-500 via-pink-500 to-amber-500 text-white shadow-md shadow-rose-500/20 group-hover:scale-105 transition-transform duration-200 animate-pulse">
                    <Icon className="h-6 w-6" />
                  </div>
                  
                  {/* Text Container with Meta Context */}
                  <div className="flex flex-col gap-0.5">
                    <span className="text-[10px] font-extrabold tracking-wider text-amber-500 uppercase">
                      {item.highlight}
                    </span>
                    <span className="text-base font-black tracking-wide text-slate-800 sm:text-lg">
                      {item.text}
                    </span>
                  </div>

                  {/* High-Tier Interactive Coupon Tag */}
                  <button 
                    onClick={() => navigator.clipboard.writeText(item.code)}
                    className="px-4 py-2 font-mono text-sm font-black tracking-widest text-rose-600 border-2 border-dashed border-rose-400 bg-gradient-to-b from-rose-50/60 to-rose-50/10 rounded-xl hover:bg-rose-600 hover:text-white hover:border-solid transition-all duration-200 active:scale-95 shadow-inner"
                  >
                    {item.code}
                  </button>
                </div>
              )
            })}
          </motion.div>
        </div>
      </div>
    </section>
  )
}