// Curated option lists for the onboarding / edit-profile UI.
//
// These are UI conveniences, not a contract: the backend resolves whatever the
// shopper picks against its own canonical vocabulary (models.ParseStyle /
// ParseCategory) when building the scoring context, and silently drops anything
// it doesn't recognise. So the labels below are chosen to (a) read well to a
// shopper and (b) map cleanly onto the engine's style/category vocabulary and
// the catalogue's data, which is what makes the resulting boosts land.

import type { Gender } from '@/types'

export const GENDER_OPTIONS: Gender[] = ['Women', 'Men', 'Unisex']

// Styles chosen to resolve against the engine's StyleTag vocabulary.
export const STYLE_OPTIONS: string[] = [
  'Ethnic',
  'Western',
  'Formal',
  'Casual',
  'Smart Casual',
  'Streetwear',
  'Athleisure',
  'Indo-Western',
  'Traditional',
  'Contemporary',
  'Festive',
  'Elegant',
  'Minimal',
  'Bohemian',
]

// Categories that map onto the catalogue's canonical category names.
export const CATEGORY_OPTIONS: string[] = [
  'Kurta Sets',
  'Sarees',
  'Ethnic Wear',
  'Dresses',
  'Tops',
  'Shirts',
  'Jeans',
  'Blazers',
  'Footwear',
  'Accessories',
]

// Colours matched loosely (case-insensitive substring) against product colour.
export const COLOR_OPTIONS: string[] = [
  'Black',
  'White',
  'Blue',
  'Navy',
  'Red',
  'Maroon',
  'Green',
  'Pink',
  'Yellow',
  'Beige',
  'Grey',
  'Pastel',
  'Gold',
  'Floral',
]

// Cities the location signal recognises (a soft climate/culture prior). Free
// text is allowed too — an unknown city simply contributes no location boost.
export const CITY_OPTIONS: string[] = [
  'Hyderabad',
  'Bengaluru',
  'Chennai',
  'Mumbai',
  'Pune',
  'Delhi',
  'New Delhi',
  'Gurgaon',
  'Noida',
  'Kolkata',
  'Ahmedabad',
  'Jaipur',
  'Lucknow',
  'Chandigarh',
  'Goa',
]

export interface BudgetBand {
  label: string
  min: number
  max: number
}

// Everyday price envelope, in whole rupees. This is context only — a per-prompt
// budget always overrides it and remains the hard filter.
export const BUDGET_BANDS: BudgetBand[] = [
  { label: 'Under ₹1,500', min: 0, max: 1500 },
  { label: '₹1,500 – ₹3,000', min: 1500, max: 3000 },
  { label: '₹3,000 – ₹6,000', min: 3000, max: 6000 },
  { label: '₹6,000 – ₹12,000', min: 6000, max: 12000 },
  { label: '₹12,000+', min: 12000, max: 100000 },
]

// Fallback brand list used only if the live /brands catalogue fetch fails, so
// the brand step is never empty. Prefer the fetched brands: they are guaranteed
// to exist in the catalogue, which makes the preferred-brand boost land.
export const FALLBACK_BRANDS: string[] = [
  'Libas',
  'Biba',
  'W',
  'Fabindia',
  'Manyavar',
  'Anouk',
  'Sangria',
  'Levis',
  'Roadster',
  'HRX',
  'H&M',
  'Nike',
  'Puma',
  'Titan',
  'Lavie',
]
