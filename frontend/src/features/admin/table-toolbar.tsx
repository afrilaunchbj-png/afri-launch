import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { Search } from "lucide-react"

import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

export interface FilterSelect {
  value: string
  onChange: (value: string) => void
  placeholder: string
  options: { value: string; label: string }[]
}

interface TableToolbarProps {
  search?: string
  onSearch?: (value: string) => void
  searchPlaceholder?: string
  selects?: FilterSelect[]
}

function SearchInput({
  value,
  onChange,
  placeholder,
}: {
  value: string
  onChange: (value: string) => void
  placeholder?: string
}) {
  const [local, setLocal] = useState(value)

  useEffect(() => setLocal(value), [value])
  useEffect(() => {
    const timer = setTimeout(() => {
      if (local !== value) onChange(local)
    }, 400)
    return () => clearTimeout(timer)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [local])

  return (
    <div className="relative w-full sm:w-72">
      <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        value={local}
        onChange={(e) => setLocal(e.target.value)}
        placeholder={placeholder}
        className="pl-9"
      />
    </div>
  )
}

/** Barre de filtres des tableaux admin : recherche + listes déroulantes. */
export function TableToolbar({ search, onSearch, searchPlaceholder, selects }: TableToolbarProps) {
  const { t } = useTranslation()
  const hasContent = Boolean(selects?.length) || Boolean(onSearch)

  if (!hasContent) return null

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
      {onSearch && (
        <SearchInput
          value={search ?? ""}
          onChange={onSearch}
          placeholder={searchPlaceholder ?? t("admin:searchPlaceholder")}
        />
      )}
      {selects?.map((s) => (
        <Select key={s.placeholder} value={s.value || "all"} onValueChange={s.onChange}>
          <SelectTrigger className="w-full sm:w-52">
            <SelectValue placeholder={s.placeholder} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{s.placeholder}</SelectItem>
            {s.options.map((o) => (
              <SelectItem key={o.value} value={o.value}>
                {o.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      ))}
    </div>
  )
}
