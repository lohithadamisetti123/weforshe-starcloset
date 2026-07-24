import type { AuthUser } from '@/types'

// authz centralises StarCloset's client-side authorization rules. Every partner
// gate — route guards, hidden buttons — reads from here rather than repeating a
// role-string comparison, so "who counts as a partner" is defined in exactly one
// place and cannot drift between call sites. This mirrors the backend guard
// (middleware.RequireRoles / models.CanManagePersonas); the server remains the
// real enforcement point, and these checks only decide what to render.

// Roles permitted to manage stylist personas. Compared case-insensitively so a
// backend value of "Partner"/"Admin" matches regardless of casing.
export const PARTNER_ROLES: readonly string[] = ['partner', 'admin']

// hasPartnerAccess reports whether a role may reach the partner portal.
export function hasPartnerAccess(role: string | null | undefined): boolean {
  return !!role && PARTNER_ROLES.includes(role.toLowerCase())
}

// canManagePersonas is the user-level convenience over hasPartnerAccess.
export function canManagePersonas(user: AuthUser | null | undefined): boolean {
  return hasPartnerAccess(user?.role)
}
