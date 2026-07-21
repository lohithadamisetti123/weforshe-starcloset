# Myntra Bharat AI

**Your Personal AI Fashion Stylist** — built for Myntra WeForShe HackerRamp 2026.

A premium, AI-first shopping assistant concept: chat with an AI stylist, get ranked outfit
recommendations with AI match scores, browse a live recommendations dashboard, and review
your search history — all running on mocked data and a mocked AI response engine (no backend
required).

## Tech stack

- React 19 + TypeScript + Vite
- Tailwind CSS v4
- Framer Motion (animation)
- Lucide React (icons)
- React Router
- TanStack React Query (provider wired in, ready for real data fetching)
- React Hook Form + Zod (installed, ready for form validation)
- Recharts (dashboard charts)

## Getting started

\`\`\`bash
npm install
npm run dev
\`\`\`

Open the printed local URL (usually http://localhost:5173).

To build for production:

\`\`\`bash
npm run build
npm run preview
\`\`\`

## Project structure

\`\`\`
src/
  components/
    layout/          Navbar, Footer, page Layout wrapper
    home/            Hero, floating cards, features, how-it-works
    assistant/       Chat window, bubbles, prompt input, AI loader, response card
    recommendations/ Product cards, dashboard stat cards, charts
    history/         History table (search, filter, pagination)
    ui/              Reusable primitives: Button, Badge, Avatar, Card, Skeleton
  pages/             One component per route
  data/              Static product catalog (40 items), mock AI engine, history, prompts
  hooks/             useChat, useTypingAnimation
  types/             Shared TypeScript interfaces
  constants/         Nav links and shared constants
  lib/               utils.ts (cn, currency/date formatting)
\`\`\`

## Notes

- All product images are placeholder photography from Picsum, seeded per product for
  consistency across reloads.
- The AI assistant is fully mocked: `src/data/mockAI.ts` parses the prompt for occasion,
  style, budget, and gender keywords and scores the catalog against them client-side.
- No backend integration — everything runs on static/mock data and local component state.
