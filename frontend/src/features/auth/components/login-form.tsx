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
import { useLogin } from "@/features/auth/hooks"
import { AppError } from "@/lib/errors"

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(1),
})

type LoginValues = z.infer<typeof loginSchema>

export function LoginForm() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const login = useLogin()

  const form = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  })

  const onSubmit = (values: LoginValues) => {
    login.mutate(values, {
      onSuccess: () => {
        toast.success(t("auth:loginSuccess"))
        navigate("/dashboard")
      },
      onError: (err) => {
        if (err instanceof AppError && err.kind === "unauthorized") {
          form.setError("root", { message: t("auth:invalidCredentials") })
        } else {
          toast.error(t("common.genericError"))
        }
      },
    })
  }

  return (
    <Card className="w-full max-w-md">
      <CardHeader className="text-center">
        <CardTitle className="font-display">{t("auth:login")}</CardTitle>
        <CardDescription>{t("auth:loginSubtitle")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            {form.formState.errors.root ? (
              <p className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {form.formState.errors.root.message}
              </p>
            ) : null}

            <FormField
              control={form.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("auth:email")}</FormLabel>
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
                  <FormLabel>{t("auth:password")}</FormLabel>
                  <FormControl>
                    <Input type="password" autoComplete="current-password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <Button type="submit" size="touch" className="w-full" disabled={login.isPending}>
              {login.isPending ? <Spinner /> : null}
              {t("auth:login")}
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
          {t("auth:noAccount")}{" "}
          <Link to="/register" className="font-medium text-primary hover:underline">
            {t("auth:createAccount")}
          </Link>
        </p>
      </CardContent>
    </Card>
  )
}
