import { createBrowserRouter } from "react-router"

import RootLayout from "@/app/root-layout"
import RootProviders from "@/app/root-providers"
import AppLayout from "@/components/layout/app-layout"
import AuthLayout from "@/components/layout/auth-layout"
import { ProtectedRoute } from "@/features/auth"
import AuthViewPage from "@/pages/auth-view"
import CreditsPage from "@/pages/credits"
import DashboardPage from "@/pages/dashboard"
import HomePage from "@/pages/home"
import IdeasPage from "@/pages/ideas"
import LoginPage from "@/pages/login"
import NotFoundPage from "@/pages/not-found"
import OpportunitiesPage from "@/pages/opportunities"
import ProjectPage from "@/pages/project"
import ProjectsPage from "@/pages/projects"
import RegisterPage from "@/pages/register"

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
          { path: "opportunities", element: <OpportunitiesPage /> },
          { path: "ideas", element: <IdeasPage /> },
          { path: "projects", element: <ProjectsPage /> },
          { path: "projects/:id", element: <ProjectPage /> },
          { path: "credits", element: <CreditsPage /> },
        ],
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
])
