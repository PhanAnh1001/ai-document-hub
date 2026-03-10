// Environment variable validation for Next.js frontend
// Backend (Go) handles its own env vars separately

const requiredEnvVars = ["NEXT_PUBLIC_API_URL"] as const;

const optionalEnvVars = [
  "NEXT_PUBLIC_SENTRY_DSN",
  "SENTRY_AUTH_TOKEN",
  "SENTRY_ORG",
  "SENTRY_PROJECT",
] as const;

type RequiredEnvVar = (typeof requiredEnvVars)[number];
type OptionalEnvVar = (typeof optionalEnvVars)[number];

function getRequiredEnv(key: RequiredEnvVar): string {
  const value = process.env[key];
  if (!value) {
    throw new Error(`Missing required environment variable: ${key}`);
  }
  return value;
}

function getOptionalEnv(key: OptionalEnvVar): string | undefined {
  return process.env[key];
}

export const env = {
  apiUrl: () => getRequiredEnv("NEXT_PUBLIC_API_URL"),
  sentryDsn: () => getOptionalEnv("NEXT_PUBLIC_SENTRY_DSN"),
} as const;
