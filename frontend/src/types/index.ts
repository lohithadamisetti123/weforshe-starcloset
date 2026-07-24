// ── Vocabulary ──────────────────────────────────────────────────────────────
// These unions MUST stay byte-identical to vocabulary.go and the SQL CHECK
// constraints in migration 000002. Any value added in Go must be added here.

export type Gender = 'Women' | 'Men' | 'Girls' | 'Boys' | 'Unisex'

export type Occasion =
  // Wedding & ceremonies
  | 'Wedding' | 'Wedding Reception' | 'Reception'
  | 'Engagement' | 'Engagement Party'
  | 'Sangeet' | 'Mehendi' | 'Haldi' | 'Winter Wedding'
  // Indian festivals
  | 'Festival' | 'Diwali' | 'Navratri' | 'Karva Chauth'
  | 'Eid' | 'Christmas' | 'New Year' | 'Pongal' | 'Onam'
  | 'Durga Puja' | 'Ganesh Chaturthi' | 'Puja' | 'Temple Visit'
  | 'Spring Festival' | 'Festive Dinner' | 'Cultural Event'
  // Work & professional
  | 'Office' | 'Office Party' | 'Business Meeting'
  | 'Interview' | 'Corporate Event' | 'Conference'
  // Education
  | 'College' | 'School' | 'Graduation' | 'Farewell' | 'Freshers'
  // Everyday
  | 'Casual' | 'Daily Wear' | 'Travel' | 'Summer Vacation'
  | 'Weekend Outing' | 'Weekend Brunch' | 'Brunch'
  | 'Picnic' | 'Shopping' | 'Staycation'
  | 'Work From Home' | 'Lounge' | 'Beach' | 'Resort'
  // Social & celebrations
  | 'Party' | 'Cocktail Party' | 'Birthday Party'
  | 'Dinner Date' | 'Date Night' | 'Cafe Date' | 'Night Out'
  | 'Family Gathering' | 'Housewarming' | 'Baby Shower'
  | 'Anniversary' | 'Anniversary Dinner'
  // Fitness & sport
  | 'Gym' | 'Sports' | 'Running' | 'Yoga'
  // Kids
  | 'Kids Party' | 'Playtime'
  // Self-care
  | 'Self Care'

export type StyleTag =
  // Classic base
  | 'Ethnic' | 'Western' | 'Formal' | 'Streetwear'
  | 'Minimal' | 'Bohemian' | 'Athleisure' | 'Indo-Western'
  // Indian ethnic extensions
  | 'Traditional' | 'Contemporary Ethnic' | 'Luxury Ethnic'
  | 'Festive' | 'Elegant' | 'Casual'
  // Fit / silhouette
  | 'Smart Casual' | 'Business Casual'
  | 'Oversized' | 'Slim Fit' | 'Regular Fit' | 'Relaxed Fit'
  // Print & embellishment
  | 'Floral' | 'Printed' | 'Graphic' | 'Embroidered'
  | 'Chikankari' | 'Bandhani' | 'Mirror Work' | 'Gota Patti'
  // Weave / fabric
  | 'Jacquard' | 'Lucknowi' | 'Organza' | 'Sharara'
  | 'Gharara' | 'Angrakha' | 'Chanderi' | 'Georgette'
  | 'Chiffon' | 'Velvet' | 'Silk' | 'Cotton' | 'Linen'
  | 'Denim' | 'Leather' | 'Suede'
  // Aesthetic registers
  | 'Vintage' | 'Classic' | 'Modern' | 'Luxury' | 'Premium'
  | 'Minimalist' | 'Resort' | 'Lounge' | 'Sporty' | 'Boho'
  | 'Contemporary'
  // Home & decor
  | 'Traditional Decor' | 'Industrial' | 'Natural'
  | 'Matte' | 'Glossy' | 'Scandinavian'

export type Category =
  // Apparel
  | 'Kurta Sets' | 'Sarees' | 'Dresses' | 'Shirts'
  | 'Tops' | 'Jeans' | 'Blazers' | 'Ethnic Wear'
  // Departments / verticals
  | 'Women' | 'Men' | 'Kids Girls' | 'Kids Boys'
  | 'Footwear' | 'Accessories' | 'Beauty'

// ── Product ─────────────────────────────────────────────────────────────────

export interface Product {
  id: string
  title: string
  brand: string
  category: Category
  price: number
  mrp: number
  discount: number
  rating: number
  ratingCount: number
  image: string
  gender: Gender
  occasion: Occasion
  style: StyleTag
  color: string
  description: string
  matchPercent?: number
}

// ── Chat ────────────────────────────────────────────────────────────────────

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
  isThinking?: boolean
  response?: AIRecommendationResponse
  collectionResponse?: CollectionResponse
  collections?: CollectionResponse[]
  // stylistReply is the persona-voice narration from the backend, shown as the
  // stylist's spoken line above the collections. Empty when no persona styled it.
  stylistReply?: string
}

export interface AIRecommendationResponse {
  summary: string
  budget: string
  occasion: string
  style: string
  reason: string
  products: Product[]
  meta?: AssistantMeta
}

// AssistantMeta mirrors the backend's per-run self-report. Only the fields the
// client currently reads are typed here; `personalized` is true when the
// shopper's profile/history produced ranking boosts on this run.
export interface AssistantMeta {
  source?: string
  confidence?: number
  latencyMs?: number
  personalized?: boolean
}

// ── History ─────────────────────────────────────────────────────────────────

export interface HistoryEntry {
  id: string
  prompt: string
  date: string
  budget: string
  occasion: string
  productsRecommended: number
  status: 'Completed' | 'Saved' | 'Purchased'
}

// ── Dashboard ────────────────────────────────────────────────────────────────

export interface DashboardStats {
  recommendationsToday: number
  averageBudget: number
  mostSelectedCategory: string
  averageAIMatch: number
}

// ── Guided assistant ─────────────────────────────────────────────────────────

export interface DepartmentOption {
  name: string
  available: boolean
}

export interface CollectionItem {
  slot: string
  product: Product
}

export interface CollectionResponse {
  title: string
  items: CollectionItem[]
  totalPrice: number
  budgetLimit?: number
  skippedSlots?: string[]
  explanation: string[]
  department: string
}

// CollectionsResponse is the payload of POST /assistant/collections: the full
// set of themed, mutually-distinct collections for one department + prompt,
// returned in a single request.
export interface CollectionsResponse {
  department: string
  collections: CollectionResponse[]
  meta?: AssistantMeta
  // stylistReply is the persona-voice narration when a stylist styled the set.
  stylistReply?: string
}

// ── Personalization profile ──────────────────────────────────────────────────
// These mirror the backend dto.ProfileResponse / dto.PastOrderResponse and the
// user_profiles / past_orders tables. The preference lists are plain strings so
// the onboarding UI can offer any curated option; the backend resolves them
// against its canonical vocabulary when building the scoring context.

export interface UserProfile {
  name: string
  gender: Gender | ''
  city: string
  budgetMin?: number
  budgetMax?: number
  favoriteStyles: string[]
  favoriteColors: string[]
  favoriteBrands: string[]
  favoriteCategories: string[]
  favoriteOccasions: string[]
  sizes: string[]
  onboarded: boolean
}

// UpdateProfileRequest is the body sent to PUT /profile/preferences. Onboarding
// sends the full profile with `onboarded: true`; a later edit sends the same
// shape (the flag stays true).
export interface UpdateProfileRequest {
  name?: string
  gender?: Gender | ''
  city?: string
  budgetMin?: number
  budgetMax?: number
  favoriteStyles?: string[]
  favoriteColors?: string[]
  favoriteBrands?: string[]
  favoriteCategories?: string[]
  favoriteOccasions?: string[]
  sizes?: string[]
  onboarded?: boolean
}

export interface PastOrder {
  id: string
  title: string
  brand: string
  category: string
  image: string
  price: number
  orderedAt: string
}

// ── Authentication ───────────────────────────────────────────────────────────
// These mirror the backend dto.AuthResponse / utils.TokenPair returned by
// POST /auth/register, /auth/login and /auth/refresh.

export interface AuthUser {
  id: string
  name: string
  email: string
  role: string
  isGuest: boolean
  createdAt: string
}

export interface AuthTokens {
  accessToken: string
  refreshToken: string
  tokenType: string
  expiresIn: number
}

export interface AuthResponse {
  user: AuthUser
  tokens: AuthTokens
}

// ── AI Chat History ──────────────────────────────────────────────────────────

export interface ChatHistoryEntry {
  id: string
  conversationId: string
  timestamp: string
  department: string
  prompt: string
  collections: CollectionResponse[]
}

// ── Saved Collections ────────────────────────────────────────────────────────

export interface SavedCollection {
  id: string
  name: string
  prompt: string
  department: string
  collection: CollectionResponse
  savedAt: string
}

// ── Collection Cart ──────────────────────────────────────────────────────────

export interface CartCollection {
  id: string
  collection: CollectionResponse
  addedAt: string
}

// ── Vocabulary helpers (for filters, dropdowns, chips) ───────────────────────

export const ALL_OCCASIONS: Occasion[] = [
  'Wedding', 'Wedding Reception', 'Reception',
  'Engagement', 'Engagement Party',
  'Sangeet', 'Mehendi', 'Haldi', 'Winter Wedding',
  'Festival', 'Diwali', 'Navratri', 'Karva Chauth',
  'Eid', 'Christmas', 'New Year', 'Pongal', 'Onam',
  'Durga Puja', 'Ganesh Chaturthi', 'Puja', 'Temple Visit',
  'Spring Festival', 'Festive Dinner', 'Cultural Event',
  'Office', 'Office Party', 'Business Meeting',
  'Interview', 'Corporate Event', 'Conference',
  'College', 'School', 'Graduation', 'Farewell', 'Freshers',
  'Casual', 'Daily Wear', 'Travel', 'Summer Vacation',
  'Weekend Outing', 'Weekend Brunch', 'Brunch',
  'Picnic', 'Shopping', 'Staycation',
  'Work From Home', 'Lounge', 'Beach', 'Resort',
  'Party', 'Cocktail Party', 'Birthday Party',
  'Dinner Date', 'Date Night', 'Cafe Date', 'Night Out',
  'Family Gathering', 'Housewarming', 'Baby Shower',
  'Anniversary', 'Anniversary Dinner',
  'Gym', 'Sports', 'Running', 'Yoga',
  'Kids Party', 'Playtime',
  'Self Care',
]

export const ALL_STYLES: StyleTag[] = [
  'Ethnic', 'Western', 'Formal', 'Streetwear',
  'Minimal', 'Bohemian', 'Athleisure', 'Indo-Western',
  'Traditional', 'Contemporary Ethnic', 'Luxury Ethnic',
  'Festive', 'Elegant', 'Casual',
  'Smart Casual', 'Business Casual',
  'Oversized', 'Slim Fit', 'Regular Fit', 'Relaxed Fit',
  'Floral', 'Printed', 'Graphic', 'Embroidered',
  'Chikankari', 'Bandhani', 'Mirror Work', 'Gota Patti',
  'Jacquard', 'Lucknowi', 'Organza', 'Sharara',
  'Gharara', 'Angrakha', 'Chanderi', 'Georgette',
  'Chiffon', 'Velvet', 'Silk', 'Cotton', 'Linen',
  'Denim', 'Leather', 'Suede',
  'Vintage', 'Classic', 'Modern', 'Luxury', 'Premium',
  'Minimalist', 'Resort', 'Lounge', 'Sporty', 'Boho',
  'Contemporary',
  'Traditional Decor', 'Industrial', 'Natural',
  'Matte', 'Glossy', 'Scandinavian',
]

export const ALL_CATEGORIES: Category[] = [
  'Kurta Sets', 'Sarees', 'Dresses', 'Shirts',
  'Tops', 'Jeans', 'Blazers', 'Ethnic Wear',
  'Women', 'Men', 'Kids Girls', 'Kids Boys',
  'Footwear', 'Accessories', 'Beauty',
]

export const ALL_GENDERS: Gender[] = ['Women', 'Men', 'Girls', 'Boys', 'Unisex']

// Occasion groups for filter chips — collapsed into meaningful clusters
export const OCCASION_GROUPS: Record<string, Occasion[]> = {
  'Wedding & Ceremonies': ['Wedding', 'Wedding Reception', 'Reception', 'Engagement', 'Engagement Party', 'Sangeet', 'Mehendi', 'Haldi', 'Winter Wedding'],
  'Festivals': ['Festival', 'Diwali', 'Navratri', 'Karva Chauth', 'Eid', 'Christmas', 'New Year', 'Pongal', 'Onam', 'Durga Puja', 'Ganesh Chaturthi', 'Puja', 'Temple Visit', 'Spring Festival', 'Festive Dinner', 'Cultural Event'],
  'Work': ['Office', 'Office Party', 'Business Meeting', 'Interview', 'Corporate Event', 'Conference'],
  'College & School': ['College', 'School', 'Graduation', 'Farewell', 'Freshers'],
  'Casual & Everyday': ['Casual', 'Daily Wear', 'Work From Home', 'Lounge', 'Shopping', 'Staycation'],
  'Travel & Outdoors': ['Travel', 'Summer Vacation', 'Weekend Outing', 'Beach', 'Resort', 'Picnic'],
  'Social': ['Party', 'Cocktail Party', 'Birthday Party', 'Night Out', 'Date Night', 'Dinner Date', 'Cafe Date', 'Brunch', 'Weekend Brunch', 'Family Gathering', 'Housewarming', 'Baby Shower', 'Anniversary', 'Anniversary Dinner'],
  'Fitness': ['Gym', 'Sports', 'Running', 'Yoga'],
  'Kids': ['Kids Party', 'Playtime'],
  'Self Care': ['Self Care'],
}
