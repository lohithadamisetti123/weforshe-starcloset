package recommendation

import (
	"strings"

	"github.com/myntra/bharat-ai-backend/internal/models"
)

// occasionAffinity[wanted][product] → how well a product tagged for occasion
// `product` serves a shopper who asked for occasion `wanted`, in [0,1].
//
// Design rules:
//   - Self-affinity is always 1.0.
//   - Absence of a key means 0.0 (intentional gap, not a miss).
//   - The matrix is asymmetric: wedding→festival=0.68 but festival→wedding=0.62
//     because the wedding brief carries a higher formality floor.
//   - Occasions that are genuinely close clusters (all the wedding sub-events,
//     all the festival sub-events) share most of their row.
//
// Groupings used in the scores below:
//
//	weddingCluster  = Wedding/Reception/WeddingReception/Sangeet/Mehendi/Haldi/Engagement/WinterWedding
//	festivalCluster = Festival/Diwali/Navratri/KarvaChauth/Eid/Christmas/NewYear/Pongal/Onam/DurgaPuja/GaneshChaturthi/Puja/TempleVisit
//	officeCluster   = Office/BusinessMeeting/Interview/CorporateEvent/Conference
//	collegeCluster  = College/School/Graduation/Farewell/Freshers
//	casualCluster   = Casual/DailyWear/WeekendOuting/Shopping/Brunch/WorkFromHome/Lounge/Staycation
//	socialCluster   = Party/CocktailParty/BirthdayParty/NightOut/DateNight/DinnerDate/CafeDate/Anniversary/AnniversaryDinner
//	travelCluster   = Travel/SummerVacation/Beach/Resort/WeekendBrunch/Picnic
//	fitnessCluster  = Gym/Sports/Running/Yoga
//	kidsCluster     = KidsParty/Playtime
var occasionAffinity = map[models.Occasion]map[models.Occasion]float64{

	// ── Wedding cluster ─────────────────────────────────────────────────────
	models.OccasionWedding: {
		models.OccasionWedding: 1.00, models.OccasionWeddingReception: 0.92,
		models.OccasionReception: 0.88, models.OccasionEngagement: 0.75,
		models.OccasionEngagementParty: 0.72, models.OccasionSangeet: 0.85,
		models.OccasionMehendi: 0.80, models.OccasionHaldi: 0.70,
		models.OccasionWinterWedding: 0.90, models.OccasionFestival: 0.65,
		models.OccasionDiwali: 0.60, models.OccasionNavratri: 0.55,
		models.OccasionCulturalEvent: 0.50, models.OccasionFestiveDinner: 0.55,
		models.OccasionParty: 0.42, models.OccasionCocktailParty: 0.38,
		models.OccasionDateNight: 0.28, models.OccasionOffice: 0.10,
		models.OccasionCollege: 0.05, models.OccasionCasual: 0.04,
	},
	models.OccasionWeddingReception: {
		models.OccasionWeddingReception: 1.00, models.OccasionWedding: 0.92,
		models.OccasionReception: 0.95, models.OccasionEngagementParty: 0.70,
		models.OccasionSangeet: 0.80, models.OccasionMehendi: 0.72,
		models.OccasionWinterWedding: 0.88, models.OccasionFestival: 0.60,
		models.OccasionParty: 0.44, models.OccasionCocktailParty: 0.40,
		models.OccasionDateNight: 0.30, models.OccasionOffice: 0.08,
		models.OccasionCasual: 0.04,
	},
	models.OccasionReception: {
		models.OccasionReception: 1.00, models.OccasionWeddingReception: 0.95,
		models.OccasionWedding: 0.88, models.OccasionEngagementParty: 0.68,
		models.OccasionSangeet: 0.75, models.OccasionFestival: 0.58,
		models.OccasionParty: 0.44, models.OccasionCocktailParty: 0.40,
		models.OccasionDateNight: 0.28, models.OccasionCasual: 0.05,
	},
	models.OccasionEngagement: {
		models.OccasionEngagement: 1.00, models.OccasionEngagementParty: 0.95,
		models.OccasionWedding: 0.75, models.OccasionSangeet: 0.72,
		models.OccasionReception: 0.68, models.OccasionFestival: 0.55,
		models.OccasionParty: 0.45, models.OccasionDateNight: 0.35,
		models.OccasionOffice: 0.10, models.OccasionCasual: 0.05,
	},
	models.OccasionEngagementParty: {
		models.OccasionEngagementParty: 1.00, models.OccasionEngagement: 0.95,
		models.OccasionWedding: 0.72, models.OccasionSangeet: 0.75,
		models.OccasionReception: 0.68, models.OccasionParty: 0.50,
		models.OccasionCocktailParty: 0.45, models.OccasionFestival: 0.55,
		models.OccasionDateNight: 0.35, models.OccasionCasual: 0.05,
	},
	models.OccasionSangeet: {
		models.OccasionSangeet: 1.00, models.OccasionMehendi: 0.82,
		models.OccasionWedding: 0.85, models.OccasionHaldi: 0.70,
		models.OccasionWeddingReception: 0.80, models.OccasionEngagementParty: 0.75,
		models.OccasionNavratri: 0.60, models.OccasionParty: 0.48,
		models.OccasionFestival: 0.60, models.OccasionCasual: 0.05,
	},
	models.OccasionMehendi: {
		models.OccasionMehendi: 1.00, models.OccasionSangeet: 0.82,
		models.OccasionHaldi: 0.78, models.OccasionWedding: 0.80,
		models.OccasionFestival: 0.58, models.OccasionNavratri: 0.55,
		models.OccasionParty: 0.40, models.OccasionCasual: 0.06,
	},
	models.OccasionHaldi: {
		models.OccasionHaldi: 1.00, models.OccasionMehendi: 0.78,
		models.OccasionSangeet: 0.70, models.OccasionWedding: 0.70,
		models.OccasionFestival: 0.55, models.OccasionNavratri: 0.50,
		models.OccasionCasual: 0.08,
	},
	models.OccasionWinterWedding: {
		models.OccasionWinterWedding: 1.00, models.OccasionWedding: 0.90,
		models.OccasionWeddingReception: 0.88, models.OccasionReception: 0.85,
		models.OccasionEngagement: 0.72, models.OccasionSangeet: 0.80,
		models.OccasionFestival: 0.60, models.OccasionParty: 0.40,
		models.OccasionCasual: 0.04,
	},

	// ── Festival cluster ────────────────────────────────────────────────────
	models.OccasionFestival: {
		models.OccasionFestival: 1.00, models.OccasionDiwali: 0.88,
		models.OccasionNavratri: 0.85, models.OccasionKarvaChauth: 0.80,
		models.OccasionEid: 0.78, models.OccasionPuja: 0.80,
		models.OccasionTempleVisit: 0.72, models.OccasionDurgaPuja: 0.82,
		models.OccasionGaneshChaturthi: 0.82, models.OccasionPongal: 0.78,
		models.OccasionOnam: 0.78, models.OccasionChristmas: 0.70,
		models.OccasionNewYear: 0.68, models.OccasionSpringFestival: 0.80,
		models.OccasionCulturalEvent: 0.75, models.OccasionFestiveDinner: 0.80,
		models.OccasionWedding: 0.62, models.OccasionParty: 0.48,
		models.OccasionOffice: 0.12, models.OccasionCollege: 0.22,
		models.OccasionCasual: 0.20,
	},
	models.OccasionDiwali: {
		models.OccasionDiwali: 1.00, models.OccasionFestival: 0.88,
		models.OccasionNavratri: 0.70, models.OccasionPuja: 0.75,
		models.OccasionKarvaChauth: 0.68, models.OccasionDurgaPuja: 0.72,
		models.OccasionGaneshChaturthi: 0.72, models.OccasionFestiveDinner: 0.80,
		models.OccasionWedding: 0.58, models.OccasionParty: 0.50,
		models.OccasionCasual: 0.18,
	},
	models.OccasionNavratri: {
		models.OccasionNavratri: 1.00, models.OccasionFestival: 0.85,
		models.OccasionDiwali: 0.70, models.OccasionKarvaChauth: 0.65,
		models.OccasionPuja: 0.72, models.OccasionSangeet: 0.60,
		models.OccasionWedding: 0.55, models.OccasionParty: 0.45,
		models.OccasionCasual: 0.20,
	},
	models.OccasionKarvaChauth: {
		models.OccasionKarvaChauth: 1.00, models.OccasionFestival: 0.80,
		models.OccasionDiwali: 0.68, models.OccasionPuja: 0.75,
		models.OccasionWedding: 0.60, models.OccasionParty: 0.40,
		models.OccasionCasual: 0.15,
	},
	models.OccasionEid: {
		models.OccasionEid: 1.00, models.OccasionFestival: 0.78,
		models.OccasionDiwali: 0.60, models.OccasionParty: 0.48,
		models.OccasionWedding: 0.55, models.OccasionCasual: 0.20,
	},
	models.OccasionChristmas: {
		models.OccasionChristmas: 1.00, models.OccasionNewYear: 0.80,
		models.OccasionFestival: 0.70, models.OccasionParty: 0.60,
		models.OccasionCocktailParty: 0.55, models.OccasionDateNight: 0.40,
		models.OccasionCasual: 0.30,
	},
	models.OccasionNewYear: {
		models.OccasionNewYear: 1.00, models.OccasionChristmas: 0.80,
		models.OccasionParty: 0.72, models.OccasionCocktailParty: 0.68,
		models.OccasionFestival: 0.65, models.OccasionDateNight: 0.45,
		models.OccasionNightOut: 0.65, models.OccasionCasual: 0.20,
	},
	models.OccasionPongal: {
		models.OccasionPongal: 1.00, models.OccasionFestival: 0.78,
		models.OccasionOnam: 0.70, models.OccasionPuja: 0.72,
		models.OccasionCasual: 0.22,
	},
	models.OccasionOnam: {
		models.OccasionOnam: 1.00, models.OccasionFestival: 0.78,
		models.OccasionPongal: 0.70, models.OccasionPuja: 0.72,
		models.OccasionCasual: 0.22,
	},
	models.OccasionDurgaPuja: {
		models.OccasionDurgaPuja: 1.00, models.OccasionFestival: 0.82,
		models.OccasionPuja: 0.85, models.OccasionNavratri: 0.72,
		models.OccasionDiwali: 0.68, models.OccasionGaneshChaturthi: 0.70,
		models.OccasionCasual: 0.18,
	},
	models.OccasionGaneshChaturthi: {
		models.OccasionGaneshChaturthi: 1.00, models.OccasionFestival: 0.82,
		models.OccasionPuja: 0.85, models.OccasionDurgaPuja: 0.70,
		models.OccasionDiwali: 0.68, models.OccasionNavratri: 0.65,
		models.OccasionCasual: 0.18,
	},
	models.OccasionPuja: {
		models.OccasionPuja: 1.00, models.OccasionTempleVisit: 0.88,
		models.OccasionFestival: 0.80, models.OccasionDurgaPuja: 0.82,
		models.OccasionGaneshChaturthi: 0.80, models.OccasionDiwali: 0.72,
		models.OccasionNavratri: 0.68, models.OccasionWedding: 0.50,
		models.OccasionCasual: 0.20,
	},
	models.OccasionTempleVisit: {
		models.OccasionTempleVisit: 1.00, models.OccasionPuja: 0.88,
		models.OccasionFestival: 0.72, models.OccasionCasual: 0.30,
	},
	models.OccasionSpringFestival: {
		models.OccasionSpringFestival: 1.00, models.OccasionFestival: 0.80,
		models.OccasionNavratri: 0.65, models.OccasionParty: 0.45,
		models.OccasionCasual: 0.25,
	},
	models.OccasionFestiveDinner: {
		models.OccasionFestiveDinner: 1.00, models.OccasionFestival: 0.80,
		models.OccasionDiwali: 0.78, models.OccasionDinnerDate: 0.60,
		models.OccasionParty: 0.55, models.OccasionCocktailParty: 0.50,
		models.OccasionWedding: 0.55, models.OccasionCasual: 0.18,
	},
	models.OccasionCulturalEvent: {
		models.OccasionCulturalEvent: 1.00, models.OccasionFestival: 0.75,
		models.OccasionWedding: 0.50, models.OccasionOffice: 0.35,
		models.OccasionCollege: 0.40, models.OccasionCasual: 0.20,
	},

	// ── Office cluster ──────────────────────────────────────────────────────
	models.OccasionOffice: {
		models.OccasionOffice: 1.00, models.OccasionBusinessMeeting: 0.92,
		models.OccasionInterview: 0.88, models.OccasionCorporateEvent: 0.85,
		models.OccasionConference: 0.82, models.OccasionOfficeParty: 0.70,
		models.OccasionCollege: 0.32, models.OccasionCasual: 0.26,
		models.OccasionTravel: 0.30, models.OccasionDateNight: 0.22,
		models.OccasionFestival: 0.12, models.OccasionWedding: 0.10,
		models.OccasionParty: 0.10,
	},
	models.OccasionOfficeParty: {
		models.OccasionOfficeParty: 1.00, models.OccasionOffice: 0.70,
		models.OccasionParty: 0.65, models.OccasionCocktailParty: 0.60,
		models.OccasionCorporateEvent: 0.75, models.OccasionConference: 0.65,
		models.OccasionFestiveDinner: 0.55, models.OccasionDateNight: 0.35,
		models.OccasionCasual: 0.22,
	},
	models.OccasionBusinessMeeting: {
		models.OccasionBusinessMeeting: 1.00, models.OccasionOffice: 0.92,
		models.OccasionInterview: 0.85, models.OccasionCorporateEvent: 0.80,
		models.OccasionConference: 0.80, models.OccasionCollege: 0.28,
		models.OccasionCasual: 0.20,
	},
	models.OccasionInterview: {
		models.OccasionInterview: 1.00, models.OccasionBusinessMeeting: 0.85,
		models.OccasionOffice: 0.88, models.OccasionCorporateEvent: 0.75,
		models.OccasionConference: 0.70, models.OccasionCollege: 0.30,
		models.OccasionCasual: 0.18,
	},
	models.OccasionCorporateEvent: {
		models.OccasionCorporateEvent: 1.00, models.OccasionOffice: 0.85,
		models.OccasionBusinessMeeting: 0.80, models.OccasionConference: 0.85,
		models.OccasionOfficeParty: 0.75, models.OccasionInterview: 0.72,
		models.OccasionParty: 0.35, models.OccasionCasual: 0.20,
	},
	models.OccasionConference: {
		models.OccasionConference: 1.00, models.OccasionCorporateEvent: 0.85,
		models.OccasionOffice: 0.82, models.OccasionBusinessMeeting: 0.80,
		models.OccasionInterview: 0.70, models.OccasionCollege: 0.40,
		models.OccasionCasual: 0.22,
	},

	// ── College / education cluster ─────────────────────────────────────────
	models.OccasionCollege: {
		models.OccasionCollege: 1.00, models.OccasionSchool: 0.65,
		models.OccasionFreshers: 0.80, models.OccasionFarewell: 0.80,
		models.OccasionGraduation: 0.72, models.OccasionCasual: 0.72,
		models.OccasionTravel: 0.46, models.OccasionParty: 0.30,
		models.OccasionOffice: 0.28, models.OccasionFestival: 0.24,
		models.OccasionDateNight: 0.20, models.OccasionWedding: 0.05,
	},
	models.OccasionSchool: {
		models.OccasionSchool: 1.00, models.OccasionCollege: 0.65,
		models.OccasionCasual: 0.68, models.OccasionFreshers: 0.60,
	},
	models.OccasionGraduation: {
		models.OccasionGraduation: 1.00, models.OccasionFarewell: 0.80,
		models.OccasionCollege: 0.72, models.OccasionOffice: 0.50,
		models.OccasionParty: 0.45, models.OccasionFestival: 0.30,
	},
	models.OccasionFarewell: {
		models.OccasionFarewell: 1.00, models.OccasionGraduation: 0.80,
		models.OccasionCollege: 0.80, models.OccasionParty: 0.50,
		models.OccasionFreshers: 0.65, models.OccasionCasual: 0.35,
	},
	models.OccasionFreshers: {
		models.OccasionFreshers: 1.00, models.OccasionFarewell: 0.65,
		models.OccasionCollege: 0.80, models.OccasionParty: 0.55,
		models.OccasionCasual: 0.40,
	},

	// ── Casual / everyday cluster ───────────────────────────────────────────
	models.OccasionCasual: {
		models.OccasionCasual: 1.00, models.OccasionDailyWear: 0.95,
		models.OccasionWeekendOuting: 0.80, models.OccasionShopping: 0.75,
		models.OccasionBrunch: 0.70, models.OccasionWorkFromHome: 0.75,
		models.OccasionLounge: 0.70, models.OccasionStaycation: 0.72,
		models.OccasionCollege: 0.70, models.OccasionTravel: 0.56,
		models.OccasionPicnic: 0.72, models.OccasionDateNight: 0.30,
		models.OccasionOffice: 0.24, models.OccasionParty: 0.20,
		models.OccasionFestival: 0.18, models.OccasionWedding: 0.05,
	},
	models.OccasionDailyWear: {
		models.OccasionDailyWear: 1.00, models.OccasionCasual: 0.95,
		models.OccasionWorkFromHome: 0.78, models.OccasionLounge: 0.72,
		models.OccasionCollege: 0.65, models.OccasionShopping: 0.70,
	},
	models.OccasionWeekendOuting: {
		models.OccasionWeekendOuting: 1.00, models.OccasionCasual: 0.80,
		models.OccasionWeekendBrunch: 0.78, models.OccasionBrunch: 0.75,
		models.OccasionShopping: 0.70, models.OccasionPicnic: 0.68,
		models.OccasionTravel: 0.55, models.OccasionDateNight: 0.35,
	},
	models.OccasionWeekendBrunch: {
		models.OccasionWeekendBrunch: 1.00, models.OccasionBrunch: 0.95,
		models.OccasionWeekendOuting: 0.78, models.OccasionCasual: 0.72,
		models.OccasionDateNight: 0.40, models.OccasionCafeDate: 0.55,
	},
	models.OccasionBrunch: {
		models.OccasionBrunch: 1.00, models.OccasionWeekendBrunch: 0.95,
		models.OccasionCafeDate: 0.70, models.OccasionDateNight: 0.45,
		models.OccasionCasual: 0.68, models.OccasionWeekendOuting: 0.72,
	},
	models.OccasionPicnic: {
		models.OccasionPicnic: 1.00, models.OccasionCasual: 0.72,
		models.OccasionWeekendOuting: 0.68, models.OccasionTravel: 0.55,
		models.OccasionBeach: 0.55,
	},
	models.OccasionShopping: {
		models.OccasionShopping: 1.00, models.OccasionCasual: 0.75,
		models.OccasionWeekendOuting: 0.70, models.OccasionBrunch: 0.55,
	},
	models.OccasionStaycation: {
		models.OccasionStaycation: 1.00, models.OccasionLounge: 0.80,
		models.OccasionCasual: 0.72, models.OccasionSelfCare: 0.65,
		models.OccasionTravel: 0.45,
	},
	models.OccasionWorkFromHome: {
		models.OccasionWorkFromHome: 1.00, models.OccasionLounge: 0.75,
		models.OccasionCasual: 0.75, models.OccasionDailyWear: 0.78,
		models.OccasionOffice: 0.45,
	},
	models.OccasionLounge: {
		models.OccasionLounge: 1.00, models.OccasionWorkFromHome: 0.75,
		models.OccasionCasual: 0.70, models.OccasionDailyWear: 0.72,
		models.OccasionSelfCare: 0.68, models.OccasionStaycation: 0.72,
	},

	// ── Travel cluster ──────────────────────────────────────────────────────
	models.OccasionTravel: {
		models.OccasionTravel: 1.00, models.OccasionSummerVacation: 0.88,
		models.OccasionBeach: 0.75, models.OccasionResort: 0.72,
		models.OccasionCasual: 0.58, models.OccasionWeekendOuting: 0.55,
		models.OccasionCollege: 0.46, models.OccasionOffice: 0.28,
		models.OccasionDateNight: 0.20, models.OccasionParty: 0.16,
		models.OccasionFestival: 0.10, models.OccasionWedding: 0.05,
	},
	models.OccasionSummerVacation: {
		models.OccasionSummerVacation: 1.00, models.OccasionTravel: 0.88,
		models.OccasionBeach: 0.82, models.OccasionResort: 0.78,
		models.OccasionCasual: 0.55, models.OccasionWeekendOuting: 0.52,
	},
	models.OccasionBeach: {
		models.OccasionBeach: 1.00, models.OccasionResort: 0.85,
		models.OccasionSummerVacation: 0.82, models.OccasionTravel: 0.75,
		models.OccasionCasual: 0.55, models.OccasionLounge: 0.45,
	},
	models.OccasionResort: {
		models.OccasionResort: 1.00, models.OccasionBeach: 0.85,
		models.OccasionSummerVacation: 0.78, models.OccasionTravel: 0.72,
		models.OccasionCasual: 0.52, models.OccasionParty: 0.35,
	},

	// ── Social / dating cluster ─────────────────────────────────────────────
	models.OccasionParty: {
		models.OccasionParty: 1.00, models.OccasionCocktailParty: 0.85,
		models.OccasionBirthdayParty: 0.82, models.OccasionNightOut: 0.80,
		models.OccasionDateNight: 0.72, models.OccasionOfficeParty: 0.65,
		models.OccasionFestival: 0.50, models.OccasionWedding: 0.42,
		models.OccasionDinnerDate: 0.55, models.OccasionCollege: 0.30,
		models.OccasionCasual: 0.20, models.OccasionTravel: 0.14,
		models.OccasionOffice: 0.10,
	},
	models.OccasionCocktailParty: {
		models.OccasionCocktailParty: 1.00, models.OccasionParty: 0.85,
		models.OccasionNightOut: 0.80, models.OccasionDateNight: 0.72,
		models.OccasionDinnerDate: 0.65, models.OccasionBirthdayParty: 0.72,
		models.OccasionFestiveDinner: 0.60, models.OccasionWedding: 0.40,
		models.OccasionOfficeParty: 0.60, models.OccasionCasual: 0.18,
	},
	models.OccasionBirthdayParty: {
		models.OccasionBirthdayParty: 1.00, models.OccasionParty: 0.82,
		models.OccasionCocktailParty: 0.72, models.OccasionNightOut: 0.65,
		models.OccasionDateNight: 0.55, models.OccasionCasual: 0.28,
	},
	models.OccasionDinnerDate: {
		models.OccasionDinnerDate: 1.00, models.OccasionDateNight: 0.90,
		models.OccasionAnniversaryDinner: 0.85, models.OccasionCocktailParty: 0.65,
		models.OccasionParty: 0.55, models.OccasionCafeDate: 0.65,
		models.OccasionOffice: 0.28, models.OccasionCasual: 0.30,
	},
	models.OccasionDateNight: {
		models.OccasionDateNight: 1.00, models.OccasionDinnerDate: 0.90,
		models.OccasionAnniversaryDinner: 0.80, models.OccasionCocktailParty: 0.72,
		models.OccasionParty: 0.70, models.OccasionCafeDate: 0.68,
		models.OccasionNightOut: 0.75, models.OccasionCasual: 0.32,
		models.OccasionFestival: 0.30, models.OccasionWedding: 0.28,
		models.OccasionOffice: 0.22, models.OccasionTravel: 0.20,
	},
	models.OccasionCafeDate: {
		models.OccasionCafeDate: 1.00, models.OccasionBrunch: 0.70,
		models.OccasionDateNight: 0.68, models.OccasionDinnerDate: 0.65,
		models.OccasionWeekendBrunch: 0.68, models.OccasionCasual: 0.55,
	},
	models.OccasionNightOut: {
		models.OccasionNightOut: 1.00, models.OccasionParty: 0.80,
		models.OccasionCocktailParty: 0.80, models.OccasionDateNight: 0.75,
		models.OccasionNewYear: 0.65, models.OccasionCasual: 0.20,
	},
	models.OccasionFamilyGathering: {
		models.OccasionFamilyGathering: 1.00, models.OccasionFestival: 0.65,
		models.OccasionHousewarming: 0.65, models.OccasionCasual: 0.55,
		models.OccasionWedding: 0.48, models.OccasionParty: 0.45,
	},
	models.OccasionHousewarming: {
		models.OccasionHousewarming: 1.00, models.OccasionFamilyGathering: 0.65,
		models.OccasionParty: 0.55, models.OccasionCasual: 0.50,
		models.OccasionFestival: 0.45,
	},
	models.OccasionBabyShower: {
		models.OccasionBabyShower: 1.00, models.OccasionFamilyGathering: 0.65,
		models.OccasionParty: 0.55, models.OccasionCasual: 0.45,
	},
	models.OccasionAnniversary: {
		models.OccasionAnniversary: 1.00, models.OccasionAnniversaryDinner: 0.92,
		models.OccasionDateNight: 0.80, models.OccasionDinnerDate: 0.78,
		models.OccasionParty: 0.50, models.OccasionCasual: 0.28,
	},
	models.OccasionAnniversaryDinner: {
		models.OccasionAnniversaryDinner: 1.00, models.OccasionAnniversary: 0.92,
		models.OccasionDinnerDate: 0.85, models.OccasionDateNight: 0.80,
		models.OccasionFestiveDinner: 0.65, models.OccasionCocktailParty: 0.60,
		models.OccasionCasual: 0.22,
	},

	// ── Fitness cluster ─────────────────────────────────────────────────────
	models.OccasionGym: {
		models.OccasionGym: 1.00, models.OccasionSports: 0.80,
		models.OccasionRunning: 0.78, models.OccasionYoga: 0.65,
		models.OccasionCasual: 0.35,
	},
	models.OccasionSports: {
		models.OccasionSports: 1.00, models.OccasionGym: 0.80,
		models.OccasionRunning: 0.82, models.OccasionYoga: 0.55,
		models.OccasionCasual: 0.30,
	},
	models.OccasionRunning: {
		models.OccasionRunning: 1.00, models.OccasionSports: 0.82,
		models.OccasionGym: 0.78, models.OccasionYoga: 0.55,
		models.OccasionCasual: 0.28,
	},
	models.OccasionYoga: {
		models.OccasionYoga: 1.00, models.OccasionGym: 0.65,
		models.OccasionSports: 0.55, models.OccasionRunning: 0.55,
		models.OccasionLounge: 0.50, models.OccasionSelfCare: 0.60,
		models.OccasionCasual: 0.40,
	},

	// ── Kids cluster ────────────────────────────────────────────────────────
	models.OccasionKidsParty: {
		models.OccasionKidsParty: 1.00, models.OccasionPlaytime: 0.65,
		models.OccasionParty: 0.50, models.OccasionBirthdayParty: 0.60,
		models.OccasionFestival: 0.45, models.OccasionCasual: 0.40,
	},
	models.OccasionPlaytime: {
		models.OccasionPlaytime: 1.00, models.OccasionKidsParty: 0.65,
		models.OccasionCasual: 0.70, models.OccasionDailyWear: 0.72,
	},

	// ── Self-care ────────────────────────────────────────────────────────────
	models.OccasionSelfCare: {
		models.OccasionSelfCare: 1.00, models.OccasionLounge: 0.68,
		models.OccasionYoga: 0.60, models.OccasionStaycation: 0.65,
		models.OccasionCasual: 0.45,
	},
}

// styleAffinity[wanted][product] → how well style `product` reads for `wanted`.
var styleAffinity = map[models.StyleTag]map[models.StyleTag]float64{

	// ── Ethnic family ───────────────────────────────────────────────────────
	models.StyleEthnic: {
		models.StyleEthnic: 1.00, models.StyleTraditional: 0.88,
		models.StyleIndoWestern: 0.70, models.StyleContemporaryEthnic: 0.72,
		models.StyleLuxuryEthnic: 0.68, models.StyleFestive: 0.65,
		// Elegant now maps to Formal in the dataset — keep low so a "Formal"
		// tagged gown/dress doesn't surface for ethnic/wedding queries.
		models.StyleElegant: 0.08, models.StyleChikankari: 0.85,
		models.StyleBandhani: 0.80, models.StyleMirrorWork: 0.80,
		models.StyleGotaPatti: 0.80, models.StyleEmbroidered: 0.82,
		models.StyleJacquard: 0.75, models.StyleOrganza: 0.70,
		models.StyleSharara: 0.85, models.StyleGharara: 0.85,
		models.StyleAngrakha: 0.85, models.StyleLucknowi: 0.85,
		models.StyleChanderi: 0.80, models.StyleGeorgette: 0.75,
		models.StyleChiffon: 0.70, models.StyleVelvet: 0.70,
		models.StyleSilk: 0.80, models.StyleBohemian: 0.30,
		models.StyleMinimal: 0.10, models.StyleCasual: 0.08,
		models.StyleFormal: 0.06, models.StyleWestern: 0.05,
		models.StyleStreetwear: 0.03, models.StyleAthleisure: 0.01,
	},
	models.StyleTraditional: {
		models.StyleTraditional: 1.00, models.StyleEthnic: 0.88,
		models.StyleChikankari: 0.82, models.StyleEmbroidered: 0.80,
		models.StyleMirrorWork: 0.78, models.StyleGotaPatti: 0.78,
		models.StyleBandhani: 0.75, models.StyleJacquard: 0.72,
		models.StyleSilk: 0.75, models.StyleChanderi: 0.75,
		models.StyleLucknowi: 0.82, models.StyleSharara: 0.80,
		models.StyleGharara: 0.80, models.StyleAngrakha: 0.80,
		models.StyleIndoWestern: 0.55, models.StyleFestive: 0.68,
		models.StyleElegant: 0.55, models.StyleMinimal: 0.12,
		models.StyleWestern: 0.06,
	},
	models.StyleContemporaryEthnic: {
		models.StyleContemporaryEthnic: 1.00, models.StyleIndoWestern: 0.82,
		models.StyleEthnic: 0.72, models.StyleTraditional: 0.60,
		models.StyleLuxuryEthnic: 0.75, models.StyleElegant: 0.65,
		models.StyleFestive: 0.65, models.StyleWestern: 0.35,
		models.StyleFormal: 0.30, models.StyleMinimal: 0.25,
	},
	models.StyleLuxuryEthnic: {
		models.StyleLuxuryEthnic: 1.00, models.StyleContemporaryEthnic: 0.75,
		models.StyleEthnic: 0.68, models.StyleTraditional: 0.65,
		models.StyleIndoWestern: 0.70, models.StyleElegant: 0.72,
		models.StyleFestive: 0.68, models.StyleLuxury: 0.75,
		models.StylePremium: 0.72, models.StyleVelvet: 0.70,
		models.StyleSilk: 0.75, models.StyleJacquard: 0.70,
	},
	models.StyleIndoWestern: {
		models.StyleIndoWestern: 1.00, models.StyleEthnic: 0.70,
		models.StyleContemporaryEthnic: 0.82, models.StyleWestern: 0.45,
		models.StyleFormal: 0.30, models.StyleMinimal: 0.30,
		models.StyleBohemian: 0.26, models.StyleStreetwear: 0.10,
		models.StyleAthleisure: 0.05, models.StyleElegant: 0.55,
	},
	models.StyleFestive: {
		models.StyleFestive: 1.00, models.StyleEthnic: 0.65,
		models.StyleTraditional: 0.68, models.StyleLuxuryEthnic: 0.68,
		models.StyleContemporaryEthnic: 0.65, models.StyleEmbroidered: 0.75,
		models.StyleMirrorWork: 0.72, models.StyleGotaPatti: 0.72,
		models.StyleJacquard: 0.70, models.StyleSilk: 0.70,
		models.StyleElegant: 0.60, models.StyleWestern: 0.35,
		models.StyleCasual: 0.10,
	},
	models.StyleElegant: {
		models.StyleElegant: 1.00, models.StyleFormal: 0.70,
		models.StyleLuxury: 0.75, models.StylePremium: 0.72,
		models.StyleFestive: 0.60, models.StyleLuxuryEthnic: 0.72,
		models.StyleContemporaryEthnic: 0.65, models.StyleMinimal: 0.45,
		models.StyleWestern: 0.40, models.StyleClassic: 0.65,
		models.StyleSilk: 0.68, models.StyleVelvet: 0.65,
	},

	// ── Western family ──────────────────────────────────────────────────────
	models.StyleWestern: {
		models.StyleWestern: 1.00, models.StyleMinimal: 0.50,
		models.StyleStreetwear: 0.45, models.StyleIndoWestern: 0.44,
		models.StyleFormal: 0.40, models.StyleBohemian: 0.36,
		models.StyleAthleisure: 0.26, models.StyleEthnic: 0.06,
		models.StyleSmartCasual: 0.55, models.StyleBusinessCasual: 0.45,
		models.StyleCasual: 0.50, models.StyleModern: 0.55,
		models.StyleContemporary: 0.52, models.StyleDenim: 0.60,
	},
	models.StyleFormal: {
		models.StyleFormal: 1.00, models.StyleMinimal: 0.56,
		models.StyleWestern: 0.40, models.StyleBusinessCasual: 0.65,
		models.StyleSmartCasual: 0.50, models.StyleElegant: 0.70,
		models.StyleLuxury: 0.60, models.StylePremium: 0.58,
		models.StyleClassic: 0.65, models.StyleIndoWestern: 0.30,
		models.StyleEthnic: 0.12, models.StyleBohemian: 0.10,
		models.StyleAthleisure: 0.05, models.StyleStreetwear: 0.05,
	},
	models.StyleMinimal: {
		models.StyleMinimal: 1.00, models.StyleMinimalist: 0.95,
		models.StyleFormal: 0.56, models.StyleWestern: 0.50,
		models.StyleAthleisure: 0.30, models.StyleIndoWestern: 0.30,
		models.StyleStreetwear: 0.26, models.StyleBohemian: 0.20,
		models.StyleEthnic: 0.15, models.StyleSmartCasual: 0.50,
		models.StyleBusinessCasual: 0.45, models.StyleCasual: 0.55,
		models.StyleClassic: 0.50,
	},
	models.StyleMinimalist: {
		models.StyleMinimalist: 1.00, models.StyleMinimal: 0.95,
		models.StyleFormal: 0.55, models.StyleWestern: 0.48,
		models.StyleClassic: 0.50, models.StyleScandinavian: 0.70,
		models.StyleNatural: 0.40,
	},
	models.StyleStreetwear: {
		models.StyleStreetwear: 1.00, models.StyleAthleisure: 0.56,
		models.StyleWestern: 0.45, models.StyleMinimal: 0.26,
		models.StyleBohemian: 0.20, models.StyleIndoWestern: 0.10,
		models.StyleOversized: 0.75, models.StyleGraphic: 0.72,
		models.StyleDenim: 0.65, models.StyleCasual: 0.55,
		models.StyleFormal: 0.05, models.StyleEthnic: 0.04,
	},
	models.StyleAthleisure: {
		models.StyleAthleisure: 1.00, models.StyleStreetwear: 0.56,
		models.StyleSporty: 0.88, models.StyleMinimal: 0.30,
		models.StyleWestern: 0.26, models.StyleBohemian: 0.10,
		models.StyleLounge: 0.55, models.StyleRelaxedFit: 0.60,
		models.StyleFormal: 0.05, models.StyleIndoWestern: 0.05,
		models.StyleEthnic: 0.02,
	},
	models.StyleSporty: {
		models.StyleSporty: 1.00, models.StyleAthleisure: 0.88,
		models.StyleStreetwear: 0.50, models.StyleMinimal: 0.25,
		models.StyleWestern: 0.20, models.StyleCasual: 0.40,
	},
	models.StyleBohemian: {
		models.StyleBohemian: 1.00, models.StyleBoho: 0.95,
		models.StyleWestern: 0.36, models.StyleEthnic: 0.30,
		models.StyleIndoWestern: 0.26, models.StyleMinimal: 0.20,
		models.StyleStreetwear: 0.20, models.StyleAthleisure: 0.10,
		models.StyleFloral: 0.65, models.StylePrinted: 0.55,
		models.StyleResort: 0.60, models.StyleRelaxedFit: 0.50,
	},
	models.StyleBoho: {
		models.StyleBoho: 1.00, models.StyleBohemian: 0.95,
		models.StyleFloral: 0.65, models.StylePrinted: 0.55,
		models.StyleResort: 0.60, models.StyleWestern: 0.34,
		models.StyleEthnic: 0.28,
	},

	// ── Fit / silhouette ────────────────────────────────────────────────────
	models.StyleSmartCasual: {
		models.StyleSmartCasual: 1.00, models.StyleBusinessCasual: 0.80,
		models.StyleFormal: 0.50, models.StyleWestern: 0.55,
		models.StyleCasual: 0.55, models.StyleMinimal: 0.50,
		models.StyleModern: 0.55,
	},
	models.StyleBusinessCasual: {
		models.StyleBusinessCasual: 1.00, models.StyleSmartCasual: 0.80,
		models.StyleFormal: 0.65, models.StyleWestern: 0.50,
		models.StyleMinimal: 0.45, models.StyleElegant: 0.48,
	},
	models.StyleCasual: {
		models.StyleCasual: 1.00, models.StyleRelaxedFit: 0.75,
		models.StyleWestern: 0.50, models.StyleMinimal: 0.55,
		models.StyleBohemian: 0.40, models.StyleStreetwear: 0.45,
		models.StyleLounge: 0.65,
	},
	models.StyleOversized: {
		models.StyleOversized: 1.00, models.StyleStreetwear: 0.75,
		models.StyleRelaxedFit: 0.80, models.StyleCasual: 0.65,
		models.StyleAthleisure: 0.55, models.StyleWestern: 0.40,
	},
	models.StyleSlimFit: {
		models.StyleSlimFit: 1.00, models.StyleFormal: 0.60,
		models.StyleWestern: 0.55, models.StyleSmartCasual: 0.55,
		models.StyleModern: 0.55, models.StyleMinimal: 0.45,
	},
	models.StyleRegularFit: {
		models.StyleRegularFit: 1.00, models.StyleCasual: 0.70,
		models.StyleWestern: 0.55, models.StyleRelaxedFit: 0.65,
		models.StyleMinimal: 0.45,
	},
	models.StyleRelaxedFit: {
		models.StyleRelaxedFit: 1.00, models.StyleRegularFit: 0.65,
		models.StyleCasual: 0.75, models.StyleBohemian: 0.50,
		models.StyleLounge: 0.70, models.StyleAthleisure: 0.60,
	},

	// ── Print / embellishment ───────────────────────────────────────────────
	models.StyleFloral: {
		models.StyleFloral: 1.00, models.StylePrinted: 0.72,
		models.StyleBohemian: 0.65, models.StyleBoho: 0.65,
		models.StyleWestern: 0.45, models.StyleEthnic: 0.40,
		models.StyleCasual: 0.50,
	},
	models.StylePrinted: {
		models.StylePrinted: 1.00, models.StyleFloral: 0.72,
		models.StyleGraphic: 0.60, models.StyleBohemian: 0.55,
		models.StyleWestern: 0.45, models.StyleCasual: 0.50,
		models.StyleEthnic: 0.35,
	},
	models.StyleGraphic: {
		models.StyleGraphic: 1.00, models.StyleStreetwear: 0.72,
		models.StylePrinted: 0.60, models.StyleWestern: 0.45,
		models.StyleCasual: 0.50, models.StyleOversized: 0.65,
	},
	models.StyleEmbroidered: {
		models.StyleEmbroidered: 1.00, models.StyleChikankari: 0.80,
		models.StyleMirrorWork: 0.72, models.StyleGotaPatti: 0.70,
		models.StyleEthnic: 0.82, models.StyleTraditional: 0.80,
		models.StyleFestive: 0.75, models.StyleBandhani: 0.60,
	},
	models.StyleChikankari: {
		models.StyleChikankari: 1.00, models.StyleLucknowi: 0.92,
		models.StyleEmbroidered: 0.80, models.StyleEthnic: 0.85,
		models.StyleTraditional: 0.82, models.StyleFestive: 0.72,
		models.StyleElegant: 0.65,
	},
	models.StyleBandhani: {
		models.StyleBandhani: 1.00, models.StyleEthnic: 0.80,
		models.StyleTraditional: 0.75, models.StylePrinted: 0.55,
		models.StyleFestive: 0.70, models.StyleFloral: 0.40,
	},
	models.StyleMirrorWork: {
		models.StyleMirrorWork: 1.00, models.StyleEthnic: 0.80,
		models.StyleTraditional: 0.78, models.StyleEmbroidered: 0.72,
		models.StyleFestive: 0.75, models.StyleGotaPatti: 0.70,
		models.StyleBandhani: 0.60,
	},
	models.StyleGotaPatti: {
		models.StyleGotaPatti: 1.00, models.StyleEthnic: 0.80,
		models.StyleTraditional: 0.78, models.StyleEmbroidered: 0.70,
		models.StyleMirrorWork: 0.70, models.StyleFestive: 0.75,
	},

	// ── Weave / fabric ──────────────────────────────────────────────────────
	models.StyleJacquard: {
		models.StyleJacquard: 1.00, models.StyleEthnic: 0.75,
		models.StyleTraditional: 0.72, models.StyleSilk: 0.70,
		models.StyleFestive: 0.70, models.StyleLuxuryEthnic: 0.70,
		models.StyleElegant: 0.60,
	},
	models.StyleLucknowi: {
		models.StyleLucknowi: 1.00, models.StyleChikankari: 0.92,
		models.StyleEthnic: 0.82, models.StyleTraditional: 0.80,
		models.StyleEmbroidered: 0.78, models.StyleFestive: 0.68,
	},
	models.StyleOrganza: {
		models.StyleOrganza: 1.00, models.StyleElegant: 0.72,
		models.StyleEthnic: 0.68, models.StyleFestive: 0.70,
		models.StyleSilk: 0.65, models.StyleContemporaryEthnic: 0.60,
	},
	models.StyleSharara: {
		models.StyleSharara: 1.00, models.StyleEthnic: 0.85,
		models.StyleTraditional: 0.80, models.StyleGharara: 0.88,
		models.StyleFestive: 0.75,
	},
	models.StyleGharara: {
		models.StyleGharara: 1.00, models.StyleSharara: 0.88,
		models.StyleEthnic: 0.85, models.StyleTraditional: 0.80,
		models.StyleFestive: 0.75,
	},
	models.StyleAngrakha: {
		models.StyleAngrakha: 1.00, models.StyleEthnic: 0.85,
		models.StyleTraditional: 0.80, models.StyleIndoWestern: 0.55,
		models.StyleFestive: 0.70,
	},
	models.StyleChanderi: {
		models.StyleChanderi: 1.00, models.StyleSilk: 0.72,
		models.StyleEthnic: 0.80, models.StyleTraditional: 0.78,
		models.StyleFestive: 0.72, models.StyleElegant: 0.65,
	},
	models.StyleGeorgette: {
		models.StyleGeorgette: 1.00, models.StyleChiffon: 0.75,
		models.StyleEthnic: 0.72, models.StyleElegant: 0.65,
		models.StyleFestive: 0.65, models.StyleWestern: 0.35,
	},
	models.StyleChiffon: {
		models.StyleChiffon: 1.00, models.StyleGeorgette: 0.75,
		models.StyleElegant: 0.65, models.StyleFestive: 0.60,
		models.StyleWestern: 0.40, models.StyleBohemian: 0.35,
	},
	models.StyleVelvet: {
		models.StyleVelvet: 1.00, models.StyleLuxuryEthnic: 0.70,
		models.StyleElegant: 0.70, models.StyleLuxury: 0.68,
		models.StyleFestive: 0.65, models.StyleEthnic: 0.68,
		models.StyleTraditional: 0.65,
	},
	models.StyleSilk: {
		models.StyleSilk: 1.00, models.StyleEthnic: 0.80,
		models.StyleTraditional: 0.78, models.StyleFestive: 0.72,
		models.StyleElegant: 0.70, models.StyleLuxuryEthnic: 0.75,
		models.StyleJacquard: 0.70, models.StyleChanderi: 0.72,
	},
	models.StyleCotton: {
		models.StyleCotton: 1.00, models.StyleCasual: 0.70,
		models.StyleMinimal: 0.65, models.StyleEthnic: 0.55,
		models.StyleRelaxedFit: 0.65, models.StyleNatural: 0.70,
	},
	models.StyleLinen: {
		models.StyleLinen: 1.00, models.StyleCasual: 0.65,
		models.StyleMinimal: 0.65, models.StyleRelaxedFit: 0.65,
		models.StyleNatural: 0.75, models.StyleResort: 0.60,
		models.StyleBohemian: 0.45,
	},
	models.StyleDenim: {
		models.StyleDenim: 1.00, models.StyleStreetwear: 0.65,
		models.StyleCasual: 0.70, models.StyleWestern: 0.60,
		models.StyleOversized: 0.55, models.StyleRelaxedFit: 0.55,
	},
	models.StyleLeather: {
		models.StyleLeather: 1.00, models.StyleLuxury: 0.65,
		models.StyleFormal: 0.55, models.StyleStreetwear: 0.50,
		models.StyleWestern: 0.50, models.StyleClassic: 0.55,
	},
	models.StyleSuede: {
		models.StyleSuede: 1.00, models.StyleLeather: 0.75,
		models.StyleLuxury: 0.60, models.StyleClassic: 0.55,
		models.StyleFormal: 0.50, models.StyleWestern: 0.45,
	},

	// ── Aesthetic registers ─────────────────────────────────────────────────
	models.StyleVintage: {
		models.StyleVintage: 1.00, models.StyleClassic: 0.72,
		models.StyleWestern: 0.45, models.StyleBohemian: 0.50,
		models.StyleStreetwear: 0.35,
	},
	models.StyleClassic: {
		models.StyleClassic: 1.00, models.StyleVintage: 0.72,
		models.StyleFormal: 0.65, models.StyleElegant: 0.65,
		models.StyleMinimal: 0.50, models.StylePremium: 0.60,
	},
	models.StyleModern: {
		models.StyleModern: 1.00, models.StyleContemporary: 0.88,
		models.StyleWestern: 0.55, models.StyleMinimal: 0.50,
		models.StyleSmartCasual: 0.55,
	},
	models.StyleLuxury: {
		models.StyleLuxury: 1.00, models.StylePremium: 0.88,
		models.StyleElegant: 0.75, models.StyleLuxuryEthnic: 0.75,
		models.StyleFormal: 0.60, models.StyleClassic: 0.60,
		models.StyleVelvet: 0.68, models.StyleSilk: 0.68,
		models.StyleLeather: 0.65, models.StyleSuede: 0.60,
	},
	models.StylePremium: {
		models.StylePremium: 1.00, models.StyleLuxury: 0.88,
		models.StyleElegant: 0.70, models.StyleFormal: 0.55,
		models.StyleClassic: 0.58, models.StyleLuxuryEthnic: 0.70,
	},
	models.StyleResort: {
		models.StyleResort: 1.00, models.StyleBohemian: 0.60,
		models.StyleBoho: 0.60, models.StyleCasual: 0.55,
		models.StyleWestern: 0.45, models.StyleLinen: 0.60,
		models.StyleFloral: 0.55,
	},
	models.StyleLounge: {
		models.StyleLounge: 1.00, models.StyleCasual: 0.65,
		models.StyleRelaxedFit: 0.70, models.StyleAthleisure: 0.55,
		models.StyleMinimal: 0.50, models.StyleCotton: 0.55,
	},
	models.StyleContemporary: {
		models.StyleContemporary: 1.00, models.StyleModern: 0.88,
		models.StyleWestern: 0.52, models.StyleMinimal: 0.50,
		models.StyleContemporaryEthnic: 0.60,
	},

	// ── Home / decor aesthetics ─────────────────────────────────────────────
	models.StyleTraditionalDecor: {
		models.StyleTraditionalDecor: 1.00, models.StyleTraditional: 0.60,
		models.StyleNatural: 0.45, models.StyleClassic: 0.55,
	},
	models.StyleIndustrial: {
		models.StyleIndustrial: 1.00, models.StyleModern: 0.55,
		models.StyleMinimal: 0.50, models.StyleContemporary: 0.52,
	},
	models.StyleNatural: {
		models.StyleNatural: 1.00, models.StyleMinimalist: 0.65,
		models.StyleCotton: 0.70, models.StyleLinen: 0.75,
		models.StyleScandinavian: 0.65, models.StyleBohemian: 0.45,
	},
	models.StyleMatte: {
		models.StyleMatte: 1.00, models.StyleMinimal: 0.55,
		models.StyleModern: 0.50,
	},
	models.StyleGlossy: {
		models.StyleGlossy: 1.00, models.StyleLuxury: 0.55,
		models.StyleModern: 0.50,
	},
	models.StyleScandinavian: {
		models.StyleScandinavian: 1.00, models.StyleMinimalist: 0.70,
		models.StyleMinimal: 0.68, models.StyleNatural: 0.65,
		models.StyleModern: 0.55,
	},
}

// occasionCategoryPrior[occasion][category] → how natural category is for
// this occasion when the shopper did not name a category themselves.
var occasionCategoryPrior = map[models.Occasion]map[models.CategoryName]float64{

	// ── Wedding cluster ─────────────────────────────────────────────────────
	models.OccasionWedding: {
		// Ethnic wear dominates — this is non-negotiable for Indian weddings.
		models.CategorySarees: 1.00, models.CategoryEthnicWear: 1.00,
		models.CategoryKurtaSets: 0.95, models.CategoryAccessories: 0.65,
		models.CategoryFootwear: 0.60,
		// Western categories strongly penalised — a wedding brief must never
		// surface casual frocks, hoodies, jeans or streetwear.
		models.CategoryDresses: 0.08, models.CategoryBlazers: 0.12,
		models.CategoryShirts: 0.05, models.CategoryTops: 0.04,
		models.CategoryJeans: 0.01, models.CategoryBeauty: 0.60,
	},
	models.OccasionWeddingReception: {
		models.CategorySarees: 1.00, models.CategoryEthnicWear: 0.95,
		models.CategoryKurtaSets: 0.90, models.CategoryAccessories: 0.65,
		models.CategoryFootwear: 0.60,
		// Receptions allow slightly more latitude (gowns, fusion) but still
		// ethnic-first. Casual western stays penalised.
		models.CategoryDresses: 0.18, models.CategoryBlazers: 0.20,
		models.CategoryShirts: 0.08, models.CategoryTops: 0.06,
		models.CategoryJeans: 0.01, models.CategoryBeauty: 0.58,
	},
	models.OccasionReception:       {models.CategorySarees: 0.95, models.CategoryEthnicWear: 0.92, models.CategoryKurtaSets: 0.88, models.CategoryDresses: 0.15, models.CategoryAccessories: 0.60, models.CategoryFootwear: 0.55, models.CategoryBlazers: 0.12, models.CategoryShirts: 0.05, models.CategoryTops: 0.05, models.CategoryJeans: 0.01, models.CategoryBeauty: 0.55},
	models.OccasionEngagement:      {models.CategoryKurtaSets: 0.90, models.CategoryEthnicWear: 0.88, models.CategorySarees: 0.80, models.CategoryDresses: 0.20, models.CategoryAccessories: 0.58, models.CategoryFootwear: 0.52, models.CategoryTops: 0.06, models.CategoryJeans: 0.02, models.CategoryBeauty: 0.55},
	models.OccasionEngagementParty: {models.CategoryKurtaSets: 0.90, models.CategoryEthnicWear: 0.90, models.CategorySarees: 0.82, models.CategoryAccessories: 0.60, models.CategoryFootwear: 0.55, models.CategoryDresses: 0.20, models.CategoryTops: 0.06, models.CategoryJeans: 0.02, models.CategoryBeauty: 0.52},
	models.OccasionSangeet:         {models.CategoryKurtaSets: 0.95, models.CategoryEthnicWear: 0.92, models.CategorySarees: 0.80, models.CategoryAccessories: 0.62, models.CategoryFootwear: 0.55, models.CategoryDresses: 0.10, models.CategoryTops: 0.05, models.CategoryJeans: 0.01, models.CategoryBeauty: 0.58},
	models.OccasionMehendi:         {models.CategoryKurtaSets: 0.95, models.CategoryEthnicWear: 0.85, models.CategorySarees: 0.70, models.CategoryAccessories: 0.58, models.CategoryFootwear: 0.50, models.CategoryDresses: 0.08, models.CategoryTops: 0.05, models.CategoryJeans: 0.01, models.CategoryBeauty: 0.55},
	models.OccasionHaldi:           {models.CategoryKurtaSets: 0.95, models.CategoryEthnicWear: 0.80, models.CategoryAccessories: 0.55, models.CategoryFootwear: 0.45, models.CategoryDresses: 0.06, models.CategoryTops: 0.04, models.CategoryJeans: 0.01, models.CategoryBeauty: 0.50},
	models.OccasionWinterWedding:   {models.CategoryEthnicWear: 1.00, models.CategorySarees: 0.92, models.CategoryKurtaSets: 0.90, models.CategoryBlazers: 0.30, models.CategoryAccessories: 0.62, models.CategoryFootwear: 0.55, models.CategoryDresses: 0.08, models.CategoryTops: 0.04, models.CategoryJeans: 0.01, models.CategoryBeauty: 0.60},

	// ── Festival cluster ────────────────────────────────────────────────────
	models.OccasionFestival: {
		models.CategoryKurtaSets: 1.00, models.CategoryEthnicWear: 0.95,
		models.CategorySarees: 0.90, models.CategoryAccessories: 0.62,
		models.CategoryFootwear: 0.55,
		// Western items are inappropriate for Indian festivals.
		models.CategoryDresses: 0.10, models.CategoryShirts: 0.08,
		models.CategoryTops: 0.08, models.CategoryBlazers: 0.06,
		models.CategoryJeans: 0.04, models.CategoryBeauty: 0.55,
	},
	models.OccasionDiwali:          {models.CategoryKurtaSets: 1.00, models.CategoryEthnicWear: 0.95, models.CategorySarees: 0.88, models.CategoryAccessories: 0.58, models.CategoryFootwear: 0.48, models.CategoryBeauty: 0.58},
	models.OccasionNavratri:        {models.CategoryKurtaSets: 0.95, models.CategoryEthnicWear: 0.92, models.CategorySarees: 0.80, models.CategoryAccessories: 0.55, models.CategoryFootwear: 0.48, models.CategoryBeauty: 0.52},
	models.OccasionKarvaChauth:     {models.CategorySarees: 0.95, models.CategoryKurtaSets: 0.90, models.CategoryEthnicWear: 0.88, models.CategoryAccessories: 0.58, models.CategoryBeauty: 0.60},
	models.OccasionEid:             {models.CategoryKurtaSets: 0.92, models.CategoryEthnicWear: 0.88, models.CategorySarees: 0.78, models.CategoryAccessories: 0.55, models.CategoryBeauty: 0.55},
	models.OccasionChristmas:       {models.CategoryDresses: 0.85, models.CategoryTops: 0.70, models.CategoryBlazers: 0.60, models.CategoryAccessories: 0.55, models.CategoryFootwear: 0.52, models.CategoryBeauty: 0.55},
	models.OccasionNewYear:         {models.CategoryDresses: 0.88, models.CategoryTops: 0.70, models.CategoryBlazers: 0.62, models.CategoryAccessories: 0.58, models.CategoryFootwear: 0.55, models.CategoryBeauty: 0.60},
	models.OccasionPongal:          {models.CategoryKurtaSets: 0.95, models.CategorySarees: 0.90, models.CategoryEthnicWear: 0.85, models.CategoryAccessories: 0.50, models.CategoryBeauty: 0.50},
	models.OccasionOnam:            {models.CategorySarees: 0.95, models.CategoryKurtaSets: 0.90, models.CategoryEthnicWear: 0.85, models.CategoryAccessories: 0.50, models.CategoryBeauty: 0.50},
	models.OccasionDurgaPuja:       {models.CategoryKurtaSets: 0.95, models.CategorySarees: 0.92, models.CategoryEthnicWear: 0.90, models.CategoryAccessories: 0.55, models.CategoryBeauty: 0.55},
	models.OccasionGaneshChaturthi: {models.CategoryKurtaSets: 0.92, models.CategoryEthnicWear: 0.90, models.CategorySarees: 0.85, models.CategoryAccessories: 0.52, models.CategoryBeauty: 0.52},
	models.OccasionPuja:            {models.CategoryKurtaSets: 0.92, models.CategorySarees: 0.90, models.CategoryEthnicWear: 0.88, models.CategoryAccessories: 0.50},
	models.OccasionTempleVisit:     {models.CategoryKurtaSets: 0.90, models.CategorySarees: 0.88, models.CategoryEthnicWear: 0.85, models.CategoryAccessories: 0.45},
	models.OccasionSpringFestival:  {models.CategoryKurtaSets: 0.90, models.CategoryEthnicWear: 0.85, models.CategorySarees: 0.80, models.CategoryDresses: 0.50, models.CategoryAccessories: 0.52},
	models.OccasionFestiveDinner:   {models.CategoryEthnicWear: 0.88, models.CategorySarees: 0.85, models.CategoryKurtaSets: 0.85, models.CategoryDresses: 0.60, models.CategoryAccessories: 0.58, models.CategoryBlazers: 0.45, models.CategoryBeauty: 0.58},
	models.OccasionCulturalEvent:   {models.CategoryKurtaSets: 0.88, models.CategoryEthnicWear: 0.85, models.CategorySarees: 0.80, models.CategoryAccessories: 0.50},

	// ── Office cluster ──────────────────────────────────────────────────────
	models.OccasionOffice: {
		models.CategoryBlazers: 1.00, models.CategoryShirts: 0.95,
		models.CategoryJeans: 0.72, // smart trousers/chinos in Jeans bucket
		models.CategoryTops:  0.62, models.CategoryFootwear: 0.60,
		models.CategoryAccessories: 0.55, models.CategoryDresses: 0.45,
		models.CategoryKurtaSets: 0.40, models.CategorySarees: 0.32,
		// Ethnic wear is inappropriate for most Indian office contexts
		models.CategoryEthnicWear: 0.08, models.CategoryBeauty: 0.45,
	},
	models.OccasionOfficeParty:     {models.CategoryDresses: 0.80, models.CategoryBlazers: 0.75, models.CategoryTops: 0.72, models.CategoryAccessories: 0.60, models.CategoryFootwear: 0.55, models.CategoryKurtaSets: 0.50, models.CategoryBeauty: 0.55},
	models.OccasionBusinessMeeting: {models.CategoryBlazers: 1.00, models.CategoryShirts: 0.95, models.CategoryTops: 0.55, models.CategoryFootwear: 0.50, models.CategoryAccessories: 0.48, models.CategoryDresses: 0.42, models.CategoryKurtaSets: 0.38},
	models.OccasionInterview:       {models.CategoryBlazers: 1.00, models.CategoryShirts: 0.95, models.CategoryTops: 0.52, models.CategoryDresses: 0.42, models.CategoryFootwear: 0.50, models.CategoryAccessories: 0.45},
	models.OccasionCorporateEvent:  {models.CategoryBlazers: 0.95, models.CategoryShirts: 0.90, models.CategoryDresses: 0.55, models.CategoryTops: 0.60, models.CategoryAccessories: 0.55, models.CategoryFootwear: 0.52},
	models.OccasionConference:      {models.CategoryBlazers: 0.92, models.CategoryShirts: 0.88, models.CategoryTops: 0.58, models.CategoryDresses: 0.50, models.CategoryAccessories: 0.48, models.CategoryFootwear: 0.48},

	// ── College / education ─────────────────────────────────────────────────
	models.OccasionCollege: {
		models.CategoryJeans: 1.00, models.CategoryTops: 0.90,
		models.CategoryFootwear: 0.78, models.CategoryShirts: 0.68,
		models.CategoryDresses: 0.55, models.CategoryKurtaSets: 0.48,
		models.CategoryAccessories: 0.50, models.CategoryBlazers: 0.15,
		models.CategoryEthnicWear: 0.05, models.CategorySarees: 0.03,
	},
	models.OccasionSchool:     {models.CategoryTops: 0.85, models.CategoryJeans: 0.82, models.CategoryDresses: 0.55, models.CategoryFootwear: 0.65, models.CategoryAccessories: 0.40},
	models.OccasionGraduation: {models.CategoryDresses: 0.80, models.CategoryBlazers: 0.75, models.CategorySarees: 0.65, models.CategoryKurtaSets: 0.60, models.CategoryFootwear: 0.60, models.CategoryAccessories: 0.55},
	models.OccasionFarewell:   {models.CategoryDresses: 0.80, models.CategoryTops: 0.72, models.CategoryJeans: 0.68, models.CategoryKurtaSets: 0.60, models.CategoryBlazers: 0.50, models.CategoryAccessories: 0.50},
	models.OccasionFreshers:   {models.CategoryDresses: 0.75, models.CategoryTops: 0.80, models.CategoryJeans: 0.75, models.CategoryKurtaSets: 0.60, models.CategoryFootwear: 0.60, models.CategoryAccessories: 0.50},

	// ── Casual / everyday ───────────────────────────────────────────────────
	models.OccasionCasual: {
		models.CategoryJeans: 0.95, models.CategoryTops: 0.90,
		models.CategoryFootwear: 0.82, models.CategoryShirts: 0.82,
		models.CategoryDresses: 0.52, models.CategoryAccessories: 0.48,
		models.CategoryKurtaSets: 0.42, models.CategoryBlazers: 0.18,
		models.CategoryEthnicWear: 0.16, models.CategorySarees: 0.12,
	},
	models.OccasionDailyWear:     {models.CategoryTops: 0.90, models.CategoryJeans: 0.92, models.CategoryShirts: 0.80, models.CategoryFootwear: 0.78, models.CategoryKurtaSets: 0.50, models.CategoryDresses: 0.45},
	models.OccasionWeekendOuting: {models.CategoryJeans: 0.90, models.CategoryTops: 0.85, models.CategoryDresses: 0.60, models.CategoryShirts: 0.78, models.CategoryFootwear: 0.75, models.CategoryAccessories: 0.52},
	models.OccasionWeekendBrunch: {models.CategoryDresses: 0.82, models.CategoryTops: 0.80, models.CategoryJeans: 0.75, models.CategoryShirts: 0.70, models.CategoryAccessories: 0.55, models.CategoryFootwear: 0.68},
	models.OccasionBrunch:        {models.CategoryDresses: 0.80, models.CategoryTops: 0.80, models.CategoryJeans: 0.72, models.CategoryShirts: 0.68, models.CategoryAccessories: 0.52, models.CategoryFootwear: 0.65},
	models.OccasionPicnic:        {models.CategoryTops: 0.88, models.CategoryJeans: 0.85, models.CategoryDresses: 0.70, models.CategoryShirts: 0.78, models.CategoryFootwear: 0.72, models.CategoryAccessories: 0.45},
	models.OccasionShopping:      {models.CategoryTops: 0.88, models.CategoryJeans: 0.88, models.CategoryDresses: 0.65, models.CategoryFootwear: 0.75, models.CategoryAccessories: 0.58},
	models.OccasionStaycation:    {models.CategoryTops: 0.82, models.CategoryJeans: 0.78, models.CategoryDresses: 0.60, models.CategoryFootwear: 0.50, models.CategoryHomeEssentials: 0.70},
	models.OccasionWorkFromHome:  {models.CategoryTops: 0.85, models.CategoryJeans: 0.80, models.CategoryKurtaSets: 0.72, models.CategoryShirts: 0.70, models.CategoryDresses: 0.50},
	models.OccasionLounge:        {models.CategoryTops: 0.85, models.CategoryJeans: 0.75, models.CategoryDresses: 0.55, models.CategoryFootwear: 0.42, models.CategoryHomeEssentials: 0.55},

	// ── Travel ──────────────────────────────────────────────────────────────
	models.OccasionTravel: {
		models.CategoryFootwear: 0.95, models.CategoryJeans: 0.92,
		models.CategoryTops: 0.88, models.CategoryShirts: 0.85,
		models.CategoryDresses: 0.48, models.CategoryAccessories: 0.55,
		models.CategoryKurtaSets: 0.30, models.CategoryBlazers: 0.10,
		models.CategoryEthnicWear: 0.04, models.CategorySarees: 0.02,
	},
	models.OccasionSummerVacation: {models.CategoryDresses: 0.85, models.CategoryTops: 0.80, models.CategoryJeans: 0.75, models.CategoryShirts: 0.75, models.CategoryFootwear: 0.80, models.CategoryAccessories: 0.60},
	models.OccasionBeach:          {models.CategoryDresses: 0.90, models.CategoryTops: 0.82, models.CategoryFootwear: 0.75, models.CategoryAccessories: 0.60, models.CategoryShirts: 0.72},
	models.OccasionResort:         {models.CategoryDresses: 0.88, models.CategoryTops: 0.80, models.CategoryShirts: 0.78, models.CategoryFootwear: 0.72, models.CategoryAccessories: 0.58, models.CategoryBlazers: 0.35},

	// ── Social ──────────────────────────────────────────────────────────────
	models.OccasionParty: {
		models.CategoryDresses: 1.00, models.CategoryTops: 0.75,
		models.CategoryFootwear: 0.70, models.CategoryAccessories: 0.68,
		models.CategoryBlazers: 0.50, models.CategoryShirts: 0.48,
		models.CategoryJeans: 0.38, models.CategorySarees: 0.35,
		models.CategoryKurtaSets: 0.32, models.CategoryEthnicWear: 0.28,
		models.CategoryBeauty: 0.68,
	},
	models.OccasionCocktailParty:     {models.CategoryDresses: 0.95, models.CategoryTops: 0.68, models.CategoryBlazers: 0.55, models.CategoryFootwear: 0.62, models.CategoryAccessories: 0.65, models.CategoryBeauty: 0.65},
	models.OccasionBirthdayParty:     {models.CategoryDresses: 0.88, models.CategoryTops: 0.72, models.CategoryJeans: 0.55, models.CategoryFootwear: 0.60, models.CategoryAccessories: 0.60, models.CategoryBeauty: 0.60},
	models.OccasionDinnerDate:        {models.CategoryDresses: 1.00, models.CategoryTops: 0.72, models.CategoryShirts: 0.62, models.CategoryFootwear: 0.62, models.CategoryAccessories: 0.58, models.CategoryBlazers: 0.42, models.CategoryKurtaSets: 0.32, models.CategoryBeauty: 0.62},
	models.OccasionDateNight:         {models.CategoryDresses: 1.00, models.CategoryTops: 0.72, models.CategoryShirts: 0.62, models.CategoryFootwear: 0.62, models.CategoryAccessories: 0.58, models.CategoryBlazers: 0.42, models.CategoryJeans: 0.38, models.CategoryKurtaSets: 0.32, models.CategoryEthnicWear: 0.32, models.CategorySarees: 0.32, models.CategoryBeauty: 0.65},
	models.OccasionCafeDate:          {models.CategoryDresses: 0.82, models.CategoryTops: 0.78, models.CategoryJeans: 0.68, models.CategoryShirts: 0.65, models.CategoryAccessories: 0.55, models.CategoryFootwear: 0.60},
	models.OccasionNightOut:          {models.CategoryDresses: 0.92, models.CategoryTops: 0.70, models.CategoryBlazers: 0.55, models.CategoryFootwear: 0.65, models.CategoryAccessories: 0.62, models.CategoryBeauty: 0.68},
	models.OccasionFamilyGathering:   {models.CategoryKurtaSets: 0.78, models.CategorySarees: 0.72, models.CategoryEthnicWear: 0.70, models.CategoryDresses: 0.55, models.CategoryTops: 0.55, models.CategoryJeans: 0.45},
	models.OccasionHousewarming:      {models.CategoryKurtaSets: 0.75, models.CategoryDresses: 0.55, models.CategoryTops: 0.60, models.CategoryJeans: 0.50, models.CategoryHomeEssentials: 0.85},
	models.OccasionBabyShower:        {models.CategoryDresses: 0.78, models.CategoryTops: 0.72, models.CategoryKurtaSets: 0.68, models.CategoryAccessories: 0.55},
	models.OccasionAnniversary:       {models.CategoryDresses: 0.88, models.CategorySarees: 0.80, models.CategoryEthnicWear: 0.75, models.CategoryKurtaSets: 0.70, models.CategoryAccessories: 0.60, models.CategoryBeauty: 0.60},
	models.OccasionAnniversaryDinner: {models.CategoryDresses: 0.92, models.CategorySarees: 0.85, models.CategoryEthnicWear: 0.80, models.CategoryAccessories: 0.62, models.CategoryFootwear: 0.58, models.CategoryBeauty: 0.65},

	// ── Fitness ─────────────────────────────────────────────────────────────
	models.OccasionGym:     {models.CategoryTops: 0.95, models.CategoryJeans: 0.90, models.CategoryFootwear: 0.88, models.CategoryAccessories: 0.40},
	models.OccasionSports:  {models.CategoryTops: 0.95, models.CategoryJeans: 0.92, models.CategoryFootwear: 0.90, models.CategoryAccessories: 0.38},
	models.OccasionRunning: {models.CategoryTops: 0.92, models.CategoryJeans: 0.88, models.CategoryFootwear: 0.95, models.CategoryAccessories: 0.35},
	models.OccasionYoga:    {models.CategoryTops: 0.90, models.CategoryJeans: 0.88, models.CategoryFootwear: 0.65, models.CategoryAccessories: 0.35},

	// ── Kids ────────────────────────────────────────────────────────────────
	models.OccasionKidsParty: {
		models.CategoryKidsGirls: 0.90, models.CategoryKidsBoys: 0.90,
		models.CategoryDresses: 0.70, models.CategoryTops: 0.65,
		models.CategoryFootwear: 0.60, models.CategoryAccessories: 0.55,
	},
	models.OccasionPlaytime: {
		models.CategoryKidsGirls: 0.95, models.CategoryKidsBoys: 0.95,
		models.CategoryTops: 0.70, models.CategoryJeans: 0.68,
		models.CategoryFootwear: 0.72,
	},

	// ── Self-care ────────────────────────────────────────────────────────────
	models.OccasionSelfCare: {
		models.CategoryBeauty: 1.00, models.CategoryHomeEssentials: 0.65,
		models.CategoryTops: 0.45, models.CategoryJeans: 0.35,
		models.CategoryAccessories: 0.40,
	},
}

// ---------------------------------------------------------------------------
// Public accessors — the scoring layer never touches the maps directly,
// ensuring consistent fallback behaviour across all lookups.
// ---------------------------------------------------------------------------

// OccasionAffinity returns how well a product tagged for productOccasion
// serves a shopper who asked for wanted. Returns 0 for any pair not in the
// matrix (intentional gap — if two occasions have no meaningful overlap we
// want 0, not 0.5).
func OccasionAffinity(wanted, productOccasion models.Occasion) float64 {
	row, ok := occasionAffinity[wanted]
	if !ok {
		return 0
	}
	return row[productOccasion]
}

// StyleAffinity returns how well productStyle reads for wanted.
func StyleAffinity(wanted, productStyle models.StyleTag) float64 {
	row, ok := styleAffinity[wanted]
	if !ok {
		return 0
	}
	return row[productStyle]
}

// CategoryPrior returns how natural category is for occasion when the shopper
// did not name a category. Unknown pairs return 0.5 — absence of evidence
// must not be scored as evidence of absence, or newly added categories would
// be silently buried.
func CategoryPrior(occasion models.Occasion, category models.CategoryName) float64 {
	row, ok := occasionCategoryPrior[occasion]
	if !ok {
		return 0.5
	}
	v, ok := row[category]
	if !ok {
		return 0.5
	}
	return v
}

// GenderAffinity returns how well a product cut for productGender serves a
// shopper who asked for wanted. Now handles Girls/Boys as well as Women/Men.
func GenderAffinity(wanted, productGender models.Gender) float64 {
	switch {
	case wanted == productGender:
		return 1.0
	case productGender == models.GenderUnisex:
		return 0.80
	case wanted == models.GenderUnisex:
		return 0.65
	// Kids cross-affinity: a Girls product weakly serves a Boys shopper and
	// vice versa when Unisex products are scarce.
	case (wanted == models.GenderGirls && productGender == models.GenderBoys) ||
		(wanted == models.GenderBoys && productGender == models.GenderGirls):
		return 0.30
	// An adult women's product never serves a kids brief and vice versa.
	default:
		return 0.0
	}
}

// ---------------------------------------------------------------------------
// Colour
// ---------------------------------------------------------------------------

var colorFamilies = map[string][]string{
	"red":    {"maroon", "wine", "wine red", "crimson", "cherry", "ruby"},
	"maroon": {"maroon", "wine", "wine red"},
	"pink":   {"blush pink", "pink", "rose", "fuchsia", "magenta", "rani pink", "dusty rose"},
	"blue":   {"navy", "indigo", "sky blue", "light blue", "dark blue", "blue", "denim", "cobalt", "powder blue", "dusty blue", "ice blue"},
	"navy":   {"navy", "dark blue", "indigo"},
	"green":  {"bottle green", "emerald green", "green", "olive", "mint", "sage", "sea green", "forest green", "mint green"},
	"teal":   {"turquoise", "teal", "aqua", "cyan"},
	"black":  {"black", "black/volt", "jet black", "charcoal black"},
	"white":  {"white", "off-white", "ivory", "cream", "pearl", "pearl white"},
	"grey":   {"charcoal", "grey melange", "steel grey", "gunmetal", "grey", "gray", "slate", "charcoal grey"},
	"gray":   {"charcoal", "grey melange", "steel grey", "gunmetal", "grey", "gray", "slate"},
	"brown":  {"tan brown", "brown", "coffee", "chocolate", "camel", "mocha brown"},
	"beige":  {"sand beige", "beige", "nude", "khaki", "off-white", "ivory", "oatmeal", "champagne"},
	"nude":   {"nude", "sand beige", "beige"},
	"khaki":  {"khaki", "sand beige", "olive"},
	"gold":   {"gold", "antique gold", "rose gold", "golden", "champagne gold"},
	"silver": {"gunmetal", "steel grey", "silver"},
	"yellow": {"mustard yellow", "yellow", "mustard", "lemon", "sunflower yellow", "turmeric yellow"},
	"orange": {"rust orange", "coral", "orange", "rust", "peach"},
	"purple": {"deep purple", "purple", "lavender", "lilac", "violet", "plum", "mauve", "royal purple"},
	"pastel": {"blush pink", "sky blue", "ivory", "off-white", "nude", "mint", "lavender", "powder blue", "peach"},
}

// ColorMatch returns how well a product's colour satisfies a wanted colour term.
func ColorMatch(wanted, productColor string) float64 {
	w := strings.ToLower(strings.TrimSpace(wanted))
	p := strings.ToLower(strings.TrimSpace(productColor))
	if w == "" || p == "" {
		return 0
	}
	if w == p {
		return 1.0
	}
	if len(w) >= 3 && strings.Contains(p, w) {
		return 0.90
	}
	if fam, ok := colorFamilies[w]; ok {
		for _, member := range fam {
			if member == p || strings.Contains(p, member) {
				return 0.75
			}
		}
	}
	return 0
}

// BestColorMatch returns the strongest match across all wanted colours.
func BestColorMatch(wanted []string, productColor string) float64 {
	var best float64
	for _, w := range wanted {
		if m := ColorMatch(w, productColor); m > best {
			best = m
		}
	}
	return best
}

// ColorFamilyTerms returns the catalogue colour names belonging to a family term.
func ColorFamilyTerms(term string) []string {
	return colorFamilies[strings.ToLower(strings.TrimSpace(term))]
}

// ---------------------------------------------------------------------------
// Brand
// ---------------------------------------------------------------------------

func BrandMatch(wanted, productBrand string) float64 {
	w := normalizeBrand(wanted)
	p := normalizeBrand(productBrand)
	if w == "" || p == "" {
		return 0
	}
	if w == p {
		return 1.0
	}
	if strings.HasPrefix(p, w) || strings.HasPrefix(w, p) {
		return 0.9
	}
	if len(w) >= 3 && strings.Contains(p, w) {
		return 0.8
	}
	return 0
}

func BestBrandMatch(wanted []string, productBrand string) float64 {
	var best float64
	for _, w := range wanted {
		if m := BrandMatch(w, productBrand); m > best {
			best = m
		}
	}
	return best
}

func normalizeBrand(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Synthetic occasion label — used by the explain layer to produce short
// human-readable occasion tags without being exhaustive.
// ---------------------------------------------------------------------------
