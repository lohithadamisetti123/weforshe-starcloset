import { useEffect, useState } from 'react'
import { NavLink, Link } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { Menu, Sparkles, X, ShoppingBag, Heart } from 'lucide-react'
import { useShop } from '@/context/ShopContext'
import { navLinks } from '@/constants/nav'
import { Button } from '@/components/ui/Button'
import { cn } from '@/lib/utils'

function NavbarBadges() {
  const { savedCount, collectionCartCount, cartCount } = useShop()
  const totalCartItems = cartCount + collectionCartCount
  return (
    <div className="flex items-center gap-2">
      {savedCount > 0 && (
        <a href="/saved" className="relative flex items-center gap-1 rounded-full border border-line bg-white px-3 py-1.5 text-xs font-semibold text-ink-700 hover:border-brand-500">
          <Heart className="h-3.5 w-3.5 text-brand-500" />
          {savedCount}
        </a>
      )}
      {totalCartItems > 0 && (
        <a href="/cart" className="relative flex items-center gap-1 rounded-full bg-ink-900 px-3 py-1.5 text-xs font-semibold text-white hover:bg-brand-600">
          <ShoppingBag className="h-3.5 w-3.5" />
          {totalCartItems}
        </a>
      )}
    </div>
  )
}

export function Navbar() {
  const [scrolled, setScrolled] = useState(false)
  const [open, setOpen] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 12)
    onScroll()
    window.addEventListener('scroll', onScroll)
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <header
      className={cn(
        'fixed inset-x-0 top-0 z-50 transition-all duration-300',
        scrolled ? 'glass border-b border-line py-2.5' : 'bg-transparent py-4'
      )}
    >
      <nav className="mx-auto flex max-w-7xl items-center justify-between px-5 lg:px-8">
        <Link to="/" className="flex items-center gap-2 group">
          <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 shadow-[0_6px_16px_-4px_rgba(255,63,108,0.55)] transition-transform group-hover:scale-105">
            <Sparkles className="h-4.5 w-4.5 text-white" strokeWidth={2.5} />
          </span>
          <span className="font-display text-[17px] font-bold tracking-tight text-ink-900">
            Myntra <span className="text-brand-500">Bharat AI</span>
          </span>
        </Link>

        <div className="hidden items-center gap-1 rounded-full border border-line bg-white/70 px-1.5 py-1.5 lg:flex">
          {navLinks.map((link) => (
            <NavLink
              key={link.to}
              to={link.to}
              end={link.to === '/'}
              className={({ isActive }) =>
                cn(
                  'rounded-full px-4 py-2 text-sm font-medium transition-colors',
                  isActive ? 'bg-ink-900 text-white' : 'text-ink-700 hover:text-ink-900'
                )
              }
            >
              {link.label}
            </NavLink>
          ))}
        </div>

        <div className="hidden items-center gap-3 lg:flex">
          <NavbarBadges />
          <Link to="/assistant">
            <Button size="md">
              <Sparkles className="h-4 w-4" />
              Try AI
            </Button>
          </Link>
        </div>

        <button
          className="flex h-10 w-10 items-center justify-center rounded-full border border-line lg:hidden"
          onClick={() => setOpen((v) => !v)}
          aria-label="Toggle menu"
        >
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </nav>

      <AnimatePresence>
        {open && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: 'auto', opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.25, ease: 'easeInOut' }}
            className="overflow-hidden border-t border-line bg-white lg:hidden"
          >
            <div className="flex flex-col gap-1 px-5 py-4">
              {navLinks.map((link) => (
                <NavLink
                  key={link.to}
                  to={link.to}
                  end={link.to === '/'}
                  onClick={() => setOpen(false)}
                  className={({ isActive }) =>
                    cn(
                      'rounded-xl px-4 py-3 text-sm font-medium',
                      isActive ? 'bg-ink-900 text-white' : 'text-ink-700'
                    )
                  }
                >
                  {link.label}
                </NavLink>
              ))}
              <Link to="/assistant" onClick={() => setOpen(false)} className="mt-2">
                <Button className="w-full">
                  <Sparkles className="h-4 w-4" />
                  Try AI
                </Button>
              </Link>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </header>
  )
}
