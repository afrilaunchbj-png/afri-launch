import { useEffect, useRef, type ReactNode } from "react"

import { cn } from "@/lib/utils"

interface RevealProps {
  children: ReactNode
  className?: string
  /** Délai avant apparition (ms) — pour décaler les cartes d'une même rangée. */
  delay?: number
}

/**
 * Reveal fait apparaître son contenu en fondu-montée à l'entrée dans le
 * viewport (IntersectionObserver). Désactivé si prefers-reduced-motion.
 */
export function Reveal({ children, className, delay = 0 }: RevealProps) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      el.classList.add("is-visible")
      return
    }
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            window.setTimeout(() => entry.target.classList.add("is-visible"), delay)
            observer.unobserve(entry.target)
          }
        }
      },
      { threshold: 0.15, rootMargin: "0px 0px -40px 0px" },
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [delay])

  return (
    <div ref={ref} className={cn("reveal", className)}>
      {children}
    </div>
  )
}
