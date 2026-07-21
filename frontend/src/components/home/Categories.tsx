import { useState, useRef } from 'react'
import { motion, useInView } from 'framer-motion'
import { Sparkles } from 'lucide-react'

const CATEGORIES = [
  { id: 'women', title: "Women's Fashion", tagline: "Elegant styles curated for every moment.", img: "src/assets/images/categories/img1.png", side: 'left' },
  { id: 'men', title: "Men's Fashion", tagline: "Premium smart casual tailored to perfection.", img: "src/assets/images/categories/img2.png", side: 'left' },
  { id: 'kids', title: "Kids' Collection", tagline: "Happy premium design for little trendsetters.", img: "src/assets/images/categories/img3.png", side: 'right' },
  { id: 'beauty', title: "Beauty & Cosmetics", tagline: "Luxury skincare and modern beauty formulas.", img: "src/assets/images/categories/img4.png", side: 'right' },
  { id: 'home', title: "Luxury Home", tagline: "Modern luxury interiors and curated spaces.", img: "src/assets/images/categories/img5.png", side: 'left' },
  { id: 'footwear', title: "Premium Footwear", tagline: "Statement luxury sneakers and crafted shoes.", img: "src/assets/images/categories/img6.png", side: 'left' },
  { id: 'watches', title: "Luxury Watches", tagline: "Timeless horology craftsmanship close-up.", img: "src/assets/images/categories/img7.png", side: 'right' },
  { id: 'accessories', title: "Elite Accessories", tagline: "High-end statement designer handbags & sunglasses.", img: "src/assets/images/categories/img8.png", side: 'right' },
]

export function Categories() {
  const containerRef = useRef(null)
  // Triggers smoothly when the container top breaks the window horizon
  const isInView = useInView(containerRef, { once: true, amount: 0.05 })

  return (
    /* Fix: pt-0 completely removes the dead space below the hero component */
    <section className="relative w-full bg-white pb-24 lg:pb-36 pt-0 overflow-hidden flex flex-col items-center">
      
      {/* --- LUXURY ATMOSPHERE BACKGROUND DECORATIONS --- */}
      <div className="absolute top-0 left-[-10%] w-[500px] h-[500px] bg-gradient-to-tr from-rose-400/10 via-pink-300/5 to-transparent blur-[140px] pointer-events-none rounded-full -z-10" />
      <div className="absolute bottom-1/4 right-[-10%] w-[600px] h-[600px] bg-gradient-to-bl from-amber-400/5 via-rose-300/10 to-transparent blur-[160px] pointer-events-none rounded-full -z-10" />
      
      <div className="absolute inset-0 pointer-events-none opacity-30 select-none bg-[radial-gradient(#f43f5e_1px,transparent_1px)] [background-size:40px_40px] [mask-image:radial-gradient(ellipse_at_center,black,transparent_75%)] -z-10" />

      {/* --- SECTION HEADER --- */}
      <div className="relative w-full max-w-[1440px] px-6 lg:px-12 mt-12 mb-16 lg:mb-24 text-center z-10 flex flex-col items-center">
        <motion.span 
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5 }}
          className="text-xs font-black tracking-[0.25em] text-transparent bg-clip-text bg-gradient-to-r from-rose-500 to-amber-500 uppercase mb-4"
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
          className="mt-6 max-w-xl text-base sm:text-lg text-slate-500 font-normal leading-relaxed tracking-wide"
        >
          Explore fashion across every category with beautifully curated collections powered by AI.
        </motion.p>
      </div>

      {/* --- HIGH LEVEL SCROLL CONVERGENCE CONTAINER --- */}
      <div 
        ref={containerRef}
        className="relative w-full max-w-[1440px] px-0 sm:px-6 lg:px-12 z-10"
      >
        {/* Mobile View: Clean Native Snap Track */}
        <div className="flex sm:hidden overflow-x-auto gap-5 px-6 pb-8 scrollbar-none snap-x snap-mandatory touch-pan-x">
          {CATEGORIES.map((category, index) => (
            <div key={category.id} className="w-[82vw] flex-shrink-0 snap-center">
              <CategoryCard category={category} index={index} isInView={true} />
            </div>
          ))}
        </div>

        {/* Tablet & Desktop View: Convergent Card Reveal Matrix */}
        <div className="hidden sm:grid grid-cols-2 lg:grid-cols-4 gap-8 overflow-visible">
          {CATEGORIES.map((category, index) => (
            <CategoryCard 
              key={category.id} 
              category={category} 
              index={index} 
              isInView={isInView} 
            />
          ))}
        </div>
      </div>
    </section>
  )
}

/* --- CONVERGENT EDITORIAL CARD --- */
function CategoryCard({ category, index, isInView }: { category: any, index: number, isInView: boolean }) {
  const [isHovered, setIsHovered] = useState(false)

  // Determine horizontal entry point based on its designated side
  const xOffset = category.side === 'left' ? -120 : 120

  // High-level cinematic entry variant (Slide inward, lose blur, lock layout)
  const cardVariants = {
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
        delay: (index % 4) * 0.08 // Staggers cleanly by columns across the grid
      }
    }
  }

  return (
    <motion.div
      initial="hidden"
      animate={isInView ? "visible" : "hidden"}
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