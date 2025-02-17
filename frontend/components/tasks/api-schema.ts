import { z } from 'zod';

export const AgentSchema = z.object({
  _id: z.string(),
  uuid: z.string(),
  hostname: z.string(),
  mac_hash: z.string(),
  nickname: z.string().optional(),
  role: z.string().optional(),
  status: z.enum(['active', 'inactive', 'disconnected']),
  last_seen: z.date(),
  created_at: z.date(),
  updated_at: z.date().optional(),
});

export type Agent = z.infer<typeof AgentSchema>;