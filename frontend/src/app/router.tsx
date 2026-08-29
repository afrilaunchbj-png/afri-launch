import { createBrowserRouter } from "react-router"

import RootLayout from "@/app/root-layout"
import AppLayout from "@/components/layout/app-layout"
import AuthLayout from "@/components/layout/auth-layout"
import { ProtectedRoute } from "@/features/auth"
import CreditsPage from "@/pages/credits"
import DashboardPage from "@/pages/dashboard"
import HomePage from "@/pages/home"
import LoginPage from "@/pages/login"
import NotFoundPage from "@/pages/not-found"
import OpportunitiesPage from "@/pages/opportunities"
import RegisterPage from "@/pages/register"

export const router = createBrowserRouter([
  {
    path: "/",
    element: <RootLayout />,
    children: [{ index: true, element: <HomePage /> }],
  },
  {
    element: <AuthLayout />,
    children: [
      { path: "login", element: <LoginPage /> },
      { path: "register", element: <RegisterPage /> },
    ],
  },
  {
    element: (
      <ProtectedRoute>
        <AppLayout />
      </ProtectedRoute>
    ),
    children: [
      { path: "dashboard", element: <DashboardPage /> },
      { path: "opportunities", element: <OpportunitiesPage /> },
      { path: "credits", element: <CreditsPage /> },
    ],
  },
  { path: "*", element: <NotFoundPage /> },
])
