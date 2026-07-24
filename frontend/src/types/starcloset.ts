// StarCloset V2 department-journey types. These mirror the backend
// dto/starcloset.go wire shapes.

// StarStylist is one stylist card within a department roster. unlockAtItems is
// department-specific — the same stylist can unlock at a different threshold in
// another department.
export interface StarStylist {
  slug: string
  displayName: string
  tagline: string
  greeting: string
  portraitUrl: string
  accentFrom: string
  accentTo: string
  unlockAtItems: number
  samplePrompts: string[]
  styleSignature: string
  defaultGender: string
}

// StarDepartment is one independent shopping journey.
export interface StarDepartment {
  key: string // "Women" | "Men" | "Kids"
  label: string
  emoji: string
  // collectionsDepartment is the engine department for Women/Men; "" for Kids,
  // where the studio derives Kids - Girls / Kids - Boys from the chosen stylist.
  collectionsDepartment: string
  stylists: StarStylist[]
}

// ProgressResponse is the payload of GET/POST /starcloset/progress.
export interface ProgressResponse {
  progress: Record<string, number>
}
