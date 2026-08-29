import { Link, useNavigate } from "react-router"
import { zodResolver } from "@hookform/resolvers/zod"
import { useForm } from "react-hook-form"
import { useTranslation } from "react-i18next"
import { z } from "zod"
import { toast } from "sonner"

import { Spinner } from "@/components/states/loading-state"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { googleAuthUrl } from "@/features/auth/api"
import { useRegister } from "@/features/auth/hooks"
import { AppError, isAppError } from "@/lib/errors"

const registerSchema = z
  .object({
    full_name: z.string().min(2),
    email: z.string().email(),
    password: z.string().min(8),
    terms: z.boolean(),
  })
  .refine((v) => v.terms, { path: ["terms"], message: "required" })

type RegisterValues = z.infer<typeof registerSchema>

export function RegisterForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const register = useRegister()

  const form = useForm<RegisterValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: { full_name: "", email: "", password: "", terms: false },
  })

  const onSubmit = (values: RegisterValues) => {
    register.mutate(
      { email: values.email, password: values.password, full_name: values.full_name },
      {
        onSuccess: () => {
          toast.success(t("auth:registerSuccess"))
          navigate("/dashboard")
        },
        onError: (err) => {
          if (isAppError(err) && err.kind === "conflict") {
            form.setError("email", { message: t("auth:emailInUse") })
          } else if (err instanceof AppError && err.fields.length > 0) {
            toast.error(t("common.genericError"))
          } else {
            toast.error(t("common.genericError"))
          }
        },
      },
    )
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="text-center">
        <CardTitle className="font-display">{t("auth:register")}</CardTitle>
        <CardDescription>{t("auth:registerSubtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="full_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("auth:fullName")} *</FormLabel>
                  <FormControl>
                    <Input autoComplete="name" placeholder="John Doe" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("auth:email")} *</FormLabel>
                  <FormControl>
                    <Input type="email" autoComplete="email" placeholder="nom@entreprise.com" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("auth:password")} *</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="new-password" {...field} />
                  </FormControl>
                  <p className="text-xs text-muted-foreground">{t("auth:passwordHint")}</p>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="terms"
              render={({ field }) => (
                <FormItem>
                  <div className="flex items-start gap-2">
                    <FormControl>
                      <Checkbox checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                    <FormLabel className="font-normal leading-snug">
                      {t("auth:terms")}{" "}
                      <span className="text-primary">{t("auth:termsLink")}</span>
                    </FormLabel>
                  </div>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" size="touch" className="w-full" disabled={register.isPending}>
              {register.isPending ? <Spinner /> : null}
              {t("auth:register")}
            </Button>
          </form>
        </Form>

        <div className="relative my-6">
          <div className="absolute inset-0 flex items-center">
            <span className="w-full border-t" />
          </div>
          <div className="relative flex justify-center">
            <span className="bg-card px-2 text-xs text-muted-foreground">{t("auth:orContinueWith")}</span>
          </div>
        </div>

        <Button variant="outline" size="touch" className="w-full" asChild>
          <a href={googleAuthUrl()}>{t("auth:google")}</a>
        </Button>

        <p className="mt-6 text-center text-sm text-muted-foreground">
          {t("auth:haveAccount")}{" "}
          <Link to="/login" className="font-medium text-primary hover:underline">
            {t("auth:login")}
          </Link>
        </p>
      </CardContent>
    </Card>
  )
}
