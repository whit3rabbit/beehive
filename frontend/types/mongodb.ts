// Base interface for common MongoDB document fields
interface BaseDocument {
    _id?: string;
    created_at: Date;
    updated_at?: Date;
  }
  
  // Admin user interface
  export interface Admin extends BaseDocument {
    username: string;
    email?: string;
    password: string; // Note: This should only be used for creation/updates
  }
  
  // Agent interface
  export interface Agent extends BaseDocument {
    uuid: string;
    hostname: string;
    mac_hash: string;
    nickname?: string;
    role?: string;
    status: 'active' | 'inactive' | 'disconnected';
    last_seen: Date;
    api_key?: string;    // Only used server-side
    api_secret?: string; // Only used server-side
  }
  
  // Role interface
  export interface Role extends BaseDocument {
    name: string;
    description?: string;
    applications: string[];
    default_tasks: string[];
  }
  
  // Task interface with specific task types
  export type TaskType = 'command_shell' | 'file_operation' | 'ui_automation' | 'browser_automation';
  export type TaskStatus = 'queued' | 'running' | 'completed' | 'failed' | 'cancelled' | 'timeout';
  
  export interface TaskOutput {
    logs?: string;
    error?: string;
  }
  
  export interface Task extends BaseDocument {
    agent_id: string;
    type: TaskType;
    parameters: Record<string, unknown>;
    status: TaskStatus;
    output?: TaskOutput;
    timeout?: number;
    started_at?: Date;
  }
  
  // System log interface
  export interface SystemLog extends BaseDocument {
    timestamp: Date;
    endpoint: string;
    agent_id?: string;
    status: string;
    details?: string;
  }
  
  // Request/Response interfaces for API calls
  export interface LoginRequest {
    username: string;
    password: string;
  }
  
  export interface LoginResponse {
    token: string;
    user: Omit<Admin, 'password'>;
  }
  
  export interface CreateAgentRequest {
    hostname: string;
    mac_hash: string;
    nickname?: string;
    role?: string;
  }
  
  export interface UpdateAgentRequest {
    nickname?: string;
    role?: string;
    status?: Agent['status'];
  }
  
  export interface CreateTaskRequest {
    agent_id: string;
    type: TaskType;
    parameters: Record<string, unknown>;
    timeout?: number;
  }
  
  export interface CreateRoleRequest {
    name: string;
    description?: string;
    applications: string[];
    default_tasks: string[];
  }
  
  // Query filter interfaces
  export interface AgentFilters {
    status?: Agent['status'];
    role?: string;
    search?: string;
  }
  
  export interface TaskFilters {
    status?: TaskStatus;
    agent_id?: string;
    type?: TaskType;
    from?: string;
    to?: string;
  }
  
  export interface LogFilters {
    agent_id?: string;
    endpoint?: string;
    status?: string;
    from?: string;
    to?: string;
  }