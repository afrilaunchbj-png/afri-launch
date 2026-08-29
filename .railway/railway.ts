import { defineRailway, github, project, service } from "railway/iac";

// Infrastructure as Code — un seul fichier pour l'environnement Railway.
//
// Les secrets / valeurs liées aux domaines sont à définir dans le dashboard
// Railway (Variables), pas ici :
//   - backend : DATABASE_URL, NEON_AUTH_BASE_URL, NEON_AUTH_JWKS_URL,
//     ALLOWED_ORIGINS (= URL publique du frontend), OPENAI_API_KEY,
//     HEYGEN_API_KEY
//   - frontend : VITE_API_URL (= URL publique du backend),
//     VITE_NEON_AUTH_URL
//
// Ces variables sont injectées au build (ARG dans le Dockerfile) et au runtime.

export default defineRailway(() => {
  const backend = service("backend", {
    source: github("afrilaunchbj-png/afri-launch", { rootDirectory: "backend" }),
    healthcheck: "/healthz",
    healthcheckTimeout: 300,
    env: {
      APP_ENV: "production",
      OPENAI_MODEL_RESEARCH: "gpt-5.6-terra",
      OPENAI_MODEL_IDEATION: "gpt-5.6-luna",
      OPENAI_MODEL_IMAGE: "gpt-image-2",
      HEYGEN_API_URL: "https://api.heygen.com",
    },
  });

  const frontend = service("frontend", {
    source: github("afrilaunchbj-png/afri-launch", { rootDirectory: "frontend" }),
    healthcheck: "/",
  });

  return project("afri-launch", {
    resources: [backend, frontend],
  });
});
