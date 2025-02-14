export interface Agent {
    uuid: string;
    hostname: string;
    mac_hash: string;
    nickname?: string;
    role?: string;
    status: string;
    last_seen: string;
  }
  
  export interface TaskOutput {
    logs?: string;
    error?: string;
  }
  
  export interface Task {
    id?: string;
    agent_id: string;
    type: 'command_shell' | 'file_operation' | 'ui_automation' | 'browser_automation';
    parameters: Record<string, any>;
    status?: 'queued' | 'running' | 'completed' | 'failed' | 'cancelled' | 'timeout';
    output?: TaskOutput;
    created_at?: string;
    updated_at?: string;
    timeout?: number;
    started_at?: string;
  }
  
  export interface Role {
    id: string;
    name: string;
    description?: string;
    applications?: string[];
    default_tasks?: string[];
    created_at: string;
  }
  
  export interface TaskFilters {
    status?: Task['status'];
    agent_id?: string;
    type?: string;
    from?: string;
    to?: string;
  }
  
  export interface AgentFilters {
    status?: string;
    role?: string;
    search?: string;
  }
  