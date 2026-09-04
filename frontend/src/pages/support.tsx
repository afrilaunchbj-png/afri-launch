import { useForm } from "react-hook-form"
import { Link } from "react-router"
import { zodResolver } from "@hookform/resolvers/zod"
import { useTranslation } from "react-i18next"
import { z } from "zod"
import { Headset } from "lucide-react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Textarea } from "@/components/ui/textarea"
import { useCreateTicket, useMyTickets } from "@/features/support/hooks"

const ticketSchema = z.object({
  subject: z.string().min(3).max(150),
  message: z.string().min(10).max(2000),
})

type TicketForm = z.infer<typeof ticketSchema>

export default function SupportPage() {
  const { t } = useTranslation()
  const { data: tickets, isLoading } = useMyTickets()
  const create = useCreateTicket()

  const form = useForm<TicketForm>({
    resolver: zodResolver(ticketSchema),
    defaultValues: { subject: "", message: "" },
  })

  const onSubmit = (values: TicketForm) => {
    create.mutate(values, {
      onSuccess: () => form.reset(),
    })
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <header>
        <h1 className="flex items-center gap-2 font-display text-2xl font-bold text-primary md:text-3xl">
          <Headset className="h-6 w-6" />
          {t("support:title")}
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">{t("support:subtitle")}</p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("support:newTicket")}</CardTitle>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="subject"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("support:subject")}</FormLabel>
                    <FormControl>
                      <input
                        {...field}
                        placeholder={t("support:subjectPlaceholder")}
                        className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="message"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("support:message")}</FormLabel>
                    <FormControl>
                      <Textarea rows={5} {...field} placeholder={t("support:messagePlaceholder")} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit" size="touch" loading={create.isPending}>
                {t("support:send")}
              </Button>
            </form>
          </Form>
        </CardContent>
      </Card>

      <section>
        <h2 className="mb-3 font-display text-lg font-semibold text-primary">{t("support:history")}</h2>
        {isLoading ? null : !tickets || tickets.length === 0 ? (
          <p className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
            {t("support:empty")}
          </p>
        ) : (
          <div className="space-y-3">
            {tickets.map((ticket) => (
              <Link key={ticket.id} to={`/support/${ticket.id}`} className="block">
                <Card className="transition-shadow hover:shadow-md">
                  <CardContent className="space-y-2 p-4">
                    <div className="flex items-start justify-between gap-3">
                      <p className="font-medium">{ticket.subject}</p>
                      <Badge variant={ticket.status === "resolved" ? "success" : "secondary"}>
                        {t(`support:status.${ticket.status}`)}
                      </Badge>
                    </div>
                    <p className="whitespace-pre-wrap text-sm text-muted-foreground">{ticket.message}</p>
                    <p className="text-xs text-muted-foreground">
                      {new Date(ticket.created_at).toLocaleString()}
                    </p>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
