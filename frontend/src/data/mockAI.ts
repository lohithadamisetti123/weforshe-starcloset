import { products } from '@/data/products'
import type { AIRecommendationResponse, Occasion, Product, StyleTag } from '@/types'

const occasionKeywords: Record<Occasion, string[]> = {
  Wedding: ['wedding', 'shaadi', 'marriage', 'sangeet'],
  Office: ['office', 'work', 'formal', 'meeting', 'boardroom'],
  College: ['college', 'campus', 'university', 'student'],
  Party: ['party', 'night out', 'club'],
  Travel: ['travel', 'vacation', 'trip', 'holiday', 'resort'],
  Festival: ['festival', 'diwali', 'puja', 'navratri', 'festive'],
  Casual: ['casual', 'everyday', 'daily', 'weekend'],
  'Date Night': ['date', 'dinner date', 'romantic'],
}

const styleKeywords: Record<StyleTag, string[]> = {
  Ethnic: ['ethnic', 'traditional', 'saree', 'kurta'],
  Western: ['western', 'dress', 'modern'],
  Formal: ['formal', 'office', 'professional'],
  Streetwear: ['street', 'streetwear', 'graphic', 'urban'],
  Minimal: ['minimal', 'simple', 'clean'],
  Bohemian: ['boho', 'bohemian', 'resort'],
  Athleisure: ['gym', 'athleisure', 'sporty', 'active'],
  'Indo-Western': ['indo-western', 'fusion'],
}

function detectOccasion(prompt: string): Occasion | null {
  const lower = prompt.toLowerCase()
  for (const [occ, keywords] of Object.entries(occasionKeywords)) {
    if (keywords.some((k) => lower.includes(k))) return occ as Occasion
  }
  return null
}

function detectStyle(prompt: string): StyleTag | null {
  const lower = prompt.toLowerCase()
  for (const [style, keywords] of Object.entries(styleKeywords)) {
    if (keywords.some((k) => lower.includes(k))) return style as StyleTag
  }
  return null
}

function detectBudget(prompt: string): number | null {
  const match = prompt.match(/(\d{3,6})/)
  if (match) return parseInt(match[1], 10)
  return null
}

function detectGender(prompt: string): 'Women' | 'Men' | null {
  const lower = prompt.toLowerCase()
  if (lower.includes('women') || lower.includes('her') || lower.includes('girl')) return 'Women'
  if (lower.includes('men') || lower.includes('him') || lower.includes('guy') || lower.includes('boy'))
    return 'Men'
  return null
}

function scoreProduct(
  product: Product,
  occasion: Occasion | null,
  style: StyleTag | null,
  budget: number | null,
  gender: 'Women' | 'Men' | null
): number {
  let score = 60
  if (occasion && product.occasion === occasion) score += 20
  if (style && product.style === style) score += 12
  if (gender && (product.gender === gender || product.gender === 'Unisex')) score += 8
  if (budget) {
    if (product.price <= budget) score += 10
    else score -= Math.min(20, ((product.price - budget) / budget) * 20)
  }
  score += (product.rating - 3.5) * 4
  return Math.max(60, Math.min(99, Math.round(score)))
}

export function generateMockResponse(prompt: string): AIRecommendationResponse {
  const occasion = detectOccasion(prompt)
  const style = detectStyle(prompt)
  const budget = detectBudget(prompt)
  const gender = detectGender(prompt)

  const scored = products
    .map((p) => ({ product: p, score: scoreProduct(p, occasion, style, budget, gender) }))
    .sort((a, b) => b.score - a.score)
    .slice(0, 6)
    .map(({ product, score }) => ({ ...product, matchPercent: score }))

  const occasionLabel = occasion ?? 'everyday'
  const styleLabel = style ?? 'versatile'
  const budgetLabel = budget ? `₹${budget.toLocaleString('en-IN')}` : 'flexible budget'

  return {
    summary: `Here's a curated edit based on "${prompt.trim()}" — picked for fit, value, and how well each piece works for the occasion.`,
    budget: budgetLabel,
    occasion: occasionLabel,
    style: styleLabel,
    reason: `These pieces were shortlisted for a ${occasionLabel.toLowerCase()} occasion in a ${styleLabel.toLowerCase()} style, ranked by AI match score across fit, rating, and price fit within your ${budgetLabel}.`,
    products: scored,
  }
}
