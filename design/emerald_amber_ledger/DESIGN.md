---
name: Emerald & Amber Ledger
colors:
  surface: '#f8f9fa'
  surface-dim: '#d9dadb'
  surface-bright: '#f8f9fa'
  surface-container-lowest: '#ffffff'
  surface-container-low: '#f3f4f5'
  surface-container: '#edeeef'
  surface-container-high: '#e7e8e9'
  surface-container-highest: '#e1e3e4'
  on-surface: '#191c1d'
  on-surface-variant: '#404944'
  inverse-surface: '#2e3132'
  inverse-on-surface: '#f0f1f2'
  outline: '#707974'
  outline-variant: '#bfc9c3'
  surface-tint: '#2b6954'
  primary: '#003527'
  on-primary: '#ffffff'
  primary-container: '#064e3b'
  on-primary-container: '#80bea6'
  inverse-primary: '#95d3ba'
  secondary: '#855300'
  on-secondary: '#ffffff'
  secondary-container: '#fea619'
  on-secondary-container: '#684000'
  tertiary: '#4f1f19'
  on-tertiary: '#ffffff'
  tertiary-container: '#6b342d'
  on-tertiary-container: '#ea9e93'
  error: '#ba1a1a'
  on-error: '#ffffff'
  error-container: '#ffdad6'
  on-error-container: '#93000a'
  primary-fixed: '#b0f0d6'
  primary-fixed-dim: '#95d3ba'
  on-primary-fixed: '#002117'
  on-primary-fixed-variant: '#0b513d'
  secondary-fixed: '#ffddb8'
  secondary-fixed-dim: '#ffb95f'
  on-secondary-fixed: '#2a1700'
  on-secondary-fixed-variant: '#653e00'
  tertiary-fixed: '#ffdad5'
  tertiary-fixed-dim: '#ffb4a9'
  on-tertiary-fixed: '#380d08'
  on-tertiary-fixed-variant: '#6e372f'
  background: '#f8f9fa'
  on-background: '#191c1d'
  surface-variant: '#e1e3e4'
  surface-white: '#FFFFFF'
  text-rich-black: '#111827'
  success-emerald: '#059669'
  innovation-amber: '#D97706'
typography:
  display-lg:
    fontFamily: Lexend
    fontSize: 48px
    fontWeight: '700'
    lineHeight: 56px
    letterSpacing: -0.02em
  headline-lg:
    fontFamily: Lexend
    fontSize: 32px
    fontWeight: '600'
    lineHeight: 40px
  headline-lg-mobile:
    fontFamily: Lexend
    fontSize: 28px
    fontWeight: '600'
    lineHeight: 36px
  headline-md:
    fontFamily: Lexend
    fontSize: 24px
    fontWeight: '600'
    lineHeight: 32px
  body-lg:
    fontFamily: Inter
    fontSize: 18px
    fontWeight: '400'
    lineHeight: 28px
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: '400'
    lineHeight: 24px
  label-md:
    fontFamily: Inter
    fontSize: 14px
    fontWeight: '600'
    lineHeight: 20px
    letterSpacing: 0.01em
  label-sm:
    fontFamily: Inter
    fontSize: 12px
    fontWeight: '500'
    lineHeight: 16px
rounded:
  sm: 0.25rem
  DEFAULT: 0.5rem
  md: 0.75rem
  lg: 1rem
  xl: 1.5rem
  full: 9999px
spacing:
  base: 8px
  container-margin-mobile: 16px
  container-margin-desktop: 40px
  gutter: 16px
  touch-target-min: 48px
---

## Brand & Style

This design system is engineered for a high-trust SaaS environment catering to African entrepreneurs. The aesthetic balances **Corporate Modernism** with **Localized Warmth**, ensuring the platform feels both globally competitive and regionally relevant. 

The brand personality is **Expert, Localized, and Efficient**. To achieve this, the design system utilizes high-contrast typography for legibility, a grounded color palette symbolizing wealth and energy, and a layout philosophy that prioritizes clarity over ornamentation. The visual direction avoids clinical coldness by using subtle organic curves and warm accents, fostering a sense of partnership and reliability.

## Colors

The palette is anchored by **Deep Emerald Green**, a color chosen to represent stability, agricultural roots, and financial growth. This is complemented by **Warm Amber**, which provides a vibrant counterpoint for calls-to-action and indicators of innovation.

- **Primary (Deep Emerald):** Used for main navigation, primary actions, and brand-heavy elements.
- **Secondary (Warm Amber):** Reserved for high-priority secondary actions, progress updates, and highlighting "new" features.
- **Backgrounds:** A crisp, very light gray (`#F9FAFB`) is used for the base canvas to reduce eye strain, while pure white (`#FFFFFF`) is used for elevated surface containers.
- **Typography:** Deep near-blacks are preferred over pure black to maintain a sophisticated, modern feel.

## Typography

The typography system uses a pairing of **Lexend** for headlines and **Inter** for body text. Lexend’s expanded character widths and geometric clarity improve readability for users who may be multitasking or using lower-end mobile displays.

- **Headlines:** Use Lexend with tight tracking for a bold, authoritative look.
- **Body:** Inter provides a neutral, highly legible experience for data-heavy SaaS interfaces.
- **Scaling:** On mobile devices, display sizes are capped at 28px to ensure content remains above the fold. 
- **Hierarchy:** Use font weight rather than just color to denote hierarchy, ensuring accessibility for all users.

## Layout & Spacing

This design system follows a **Mobile-First Fluid Grid** model. The spacing rhythm is based on an 8px baseline to ensure consistency across all components.

- **Mobile:** 4-column fluid grid with 16px side margins.
- **Tablet:** 8-column fluid grid with 24px side margins.
- **Desktop:** 12-column fixed-width grid (max 1280px) centered on the screen.
- **Touch Targets:** All interactive elements must maintain a minimum height/width of 48px to accommodate one-handed mobile use, typical of entrepreneurs on the move.

## Elevation & Depth

Hierarchy is established through **Tonal Layering** and **Ambient Shadows**. This system avoids heavy borders in favor of soft, diffused shadows that simulate physical depth.

- **Level 0 (Base):** Light gray background (`#F9FAFB`).
- **Level 1 (Cards):** Pure white surfaces with a 4px blur, 2px Y-offset shadow at 5% opacity.
- **Level 2 (Dropdowns/Modals):** Pure white surfaces with a 12px blur, 6px Y-offset shadow at 10% opacity.
- **Active States:** Subtle inner shadows or color shifts in the primary Emerald green denote selection.

## Shapes

The design system uses a **Rounded** shape language (`0.5rem` or `8px` base) to strike a balance between professional structure and modern friendliness.

- **Standard Elements:** Buttons and Input fields use an 8px radius.
- **Large Containers:** Cards and Modals use a 16px radius (`rounded-lg`).
- **Status Indicators:** Pills and progress chips use full rounding (pill-shaped) to distinguish them from actionable buttons.

## Components

### Buttons
Primary buttons are solid Deep Emerald with white text. Secondary buttons use an Emerald outline. All buttons must have a minimum height of 48px for mobile accessibility.

### Cards
Cards are the primary organizational unit. They must feature a white background, 16px padding, and the Level 1 shadow defined in the Elevation section.

### Progress Indicators
Since many SaaS tasks for entrepreneurs may be asynchronous (e.g., credit checks, document processing), use large, clear progress bars in Warm Amber. Include a status label (e.g., "Verifying...") in Inter Label-MD.

### Input Fields
Inputs use a 1px border in light gray, which transitions to Deep Emerald on focus. Labels should always be visible (never placeholder-only) to ensure clarity during data entry.

### Lists
Use "Divided Lists" for mobile efficiency. Each list item should have a minimum height of 64px with a chevron icon to indicate drill-down capability.