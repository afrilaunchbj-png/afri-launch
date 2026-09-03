import { createBrowserRouter, Navigate } from "react-router"

import RootLayout from "@/app/root-layout"
import RootProviders from "@/app/root-providers"
import AppLayout from "@/components/layout/app-layout"
import AuthLayout from "@/components/layout/auth-layout"
import { ProtectedRoute } from "@/features/auth"
import AdminPage from "@/pages/admin"
import AuthViewPage from "@/pages/auth-view"
import CreditsPage from "@/pages/credits"
import DashboardPage from "@/pages/dashboard"
import DiscoverPage from "@/pages/discover"
import HomePage from "@/pages/home"
import LoginPage from "@/pages/login"
import NotFoundPage from "@/pages/not-found"
import ProjectPage from "@/pages/project"
import ProjectsPage from "@/pages/projects"
import RegisterPage from "@/pages/register"
import SettingsPage from "@/pages/settings"
import SupportPage from "@/pages/support"

export const router = createBrowserRouter([
  {
    path: "/",
    element: <RootProviders />,
    children: [
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
          { path: "auth/*", element: <AuthViewPage /> },
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
          { path: "discover", element: <DiscoverPage /> },
          // Ancien parcours (catalogue + liste d'idées) remplacé par le chat.
          { path: "opportunities", element: <Navigate to="/discover" replace /> },
          { path: "ideas", element: <Navigate to="/discover" replace /> },
          { path: "projects", element: <ProjectsPage /> },
          { path: "projects/:id", element: <ProjectPage /> },
          { path: "credits", element: <CreditsPage /> },
          { path: "support", element: <SupportPage /> },
          { path: "settings", element: <SettingsPage /> },
          { path: "admin", element: <AdminPage /> },
        ],
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])
