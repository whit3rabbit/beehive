// lib/env.ts
import { z } from 'zod';

// Environment variable schema
const envSchema = z.object({
  // Core URLs derived from environment
  MANAGER_HOST: z.string().default('localhost'),
  MANAGER_PORT: z.string().default('8080'),
  FRONTEND_HOST: z.string().default('localhost'),
  FRONTEND_PORT: z.string().default('3000'),
  
  // Optional variables with defaults
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  DOMAIN: z.string().default('localhost'),
});

// Environment type based on schema
type Env = z.infer<typeof envSchema>;

// Function to get validated environment variables
export function getEnvConfig(): Env {
  try {
    return envSchema.parse(process.env);
  } catch (error) {
    console.error('Invalid environment variables:', error);
    throw new Error('Invalid environment configuration');
  }
}

// Get environment config
export const env = getEnvConfig();

// Derived URLs based on environment
export const urls = {
  api: `http://${env.MANAGER_HOST}:${env.MANAGER_PORT}/api`,
  ws: `ws://${env.MANAGER_HOST}:${env.MANAGER_PORT}/ws`,
  frontend: `http://${env.FRONTEND_HOST}:${env.FRONTEND_PORT}`,
};

// Config object for different environments
export const config = {
  development: {
    apiUrl: urls.api,
    wsUrl: urls.ws,
    frontendUrl: urls.frontend,
  },
  production: {
    apiUrl: `https://${env.DOMAIN}/api`,
    wsUrl: `wss://${env.DOMAIN}/ws`,
    frontendUrl: `https://${env.DOMAIN}`,
  },
  test: {
    apiUrl: urls.api,
    wsUrl: urls.ws,
    frontendUrl: urls.frontend,
  },
} as const;

// Current environment configuration
export const currentConfig = config[env.NODE_ENV];