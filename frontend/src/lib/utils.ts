import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatINR(value: number): string {
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    maximumFractionDigits: 0,
  }).format(value)
}

export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('en-IN', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })
}

// A neutral placeholder shown when a product image fails to load. Prevents the
// browser from rendering overflowing `alt` text over the card (the "title
// bleeding into the image" bug when an image 404s).
export const PRODUCT_IMG_FALLBACK =
  "data:image/svg+xml;utf8," +
  encodeURIComponent(
    `<svg xmlns='http://www.w3.org/2000/svg' width='300' height='400'>` +
      `<rect width='300' height='400' fill='%23f1f1f2'/>` +
      `<g fill='none' stroke='%23bcbcc4' stroke-width='2'>` +
      `<rect x='120' y='168' width='60' height='48' rx='4'/>` +
      `<circle cx='138' cy='186' r='6'/>` +
      `<path d='M126 210l18-16 12 10 10-8 16 14'/></g></svg>`,
  )

// onError handler for any product <img>. Swaps to the placeholder once and
// clears alt so no text overflows the image box.
export function onImgError(e: { currentTarget: HTMLImageElement }) {
  const t = e.currentTarget
  if (t.dataset.fallback) return
  t.dataset.fallback = '1'
  t.src = PRODUCT_IMG_FALLBACK
  t.alt = ''
}
