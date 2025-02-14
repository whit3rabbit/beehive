// api/types.ts
export interface Agent {
  uuid: string;
  hostname: string;
  mac_hash: string;
  nickname?: string;
  role?: string;
  status?: string;
  last_seen?: string;
}
