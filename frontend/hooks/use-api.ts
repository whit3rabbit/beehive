import { useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { socket } from '@/lib/query-client';
import type {
  Agent,
  Task,
  Role,
  SystemLog,
  TaskFilters,
  AgentFilters,
  LogFilters,
  CreateTaskRequest,
  CreateAgentRequest,
  UpdateAgentRequest,
  CreateRoleRequest,
  LoginRequest,
  LoginResponse
} from '@/types/mongodb';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add authentication interceptor
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Authentication hooks
export function useLogin() {
  return useMutation({
    mutationFn: async (credentials: LoginRequest) => {
      const { data } = await api.post<LoginResponse>('/auth/login', credentials);
      localStorage.setItem('auth_token', data.token);
      return data;
    },
  });
}

export function useLogout() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await api.post('/auth/logout');
      localStorage.removeItem('auth_token');
      queryClient.clear();
    },
  });
}

// Agent hooks
export function useAgents(filters?: AgentFilters) {
    return useQuery({
      queryKey: ['agents', filters],
      queryFn: async () => {
        try {
          const { data } = await api.get<Agent[]>('/agents', { params: filters });
          return data;
        } catch (error) {
          // Log to error reporting service
          console.error('Failed to fetch agents:', error);
          throw error;
        }
      },
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
    });
  }

export function useAgent(uuid: string) {
  return useQuery({
    queryKey: ['agents', uuid],
    queryFn: async () => {
      const { data } = await api.get<Agent>(`/agents/${uuid}`);
      return data;
    },
    enabled: !!uuid,
  });
}

export function useCreateAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (newAgent: CreateAgentRequest) => {
      const { data } = await api.post<Agent>('/agents', newAgent);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] });
    },
  });
}

export function useUpdateAgent() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ uuid, data }: { uuid: string; data: UpdateAgentRequest }) => {
      const response = await api.put<Agent>(`/agents/${uuid}`, data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['agents'] });
    },
  });
}

// Task hooks
export function useTasks(filters?: TaskFilters) {
  const queryClient = useQueryClient();

  useEffect(() => {
    const handleTaskUpdate = (updatedTask: Task) => {
      queryClient.setQueryData<Task[]>(['tasks'], (old) => 
        old?.map(task => task._id === updatedTask._id ? updatedTask : task) ?? [updatedTask]
      );
    };

    socket.on('task:update', handleTaskUpdate);
    return () => {
      socket.off('task:update', handleTaskUpdate);
    };
  }, [queryClient]);

  return useQuery({
    queryKey: ['tasks', filters],
    queryFn: async () => {
      const { data } = await api.get<Task[]>('/tasks', { params: filters });
      return data;
    },
  });
}

export function useTask(id: string) {
  return useQuery({
    queryKey: ['tasks', id],
    queryFn: async () => {
      const { data } = await api.get<Task>(`/tasks/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (task: CreateTaskRequest) => {
      const { data } = await api.post<Task>('/tasks', task);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
}

export function useCancelTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      const { data } = await api.post<Task>(`/tasks/${id}/cancel`);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
    },
  });
}

// Role hooks
export function useRoles() {
  return useQuery({
    queryKey: ['roles'],
    queryFn: async () => {
      const { data } = await api.get<Role[]>('/roles');
      return data;
    },
  });
}

export function useRole(id: string) {
  return useQuery({
    queryKey: ['roles', id],
    queryFn: async () => {
      const { data } = await api.get<Role>(`/roles/${id}`);
      return data;
    },
    enabled: !!id,
  });
}

export function useCreateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (role: CreateRoleRequest) => {
      const { data } = await api.post<Role>('/roles', role);
      return data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
    },
  });
}

export function useUpdateRole() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ id, data }: { id: string; data: Partial<Role> }) => {
      const response = await api.put<Role>(`/roles/${id}`, data);
      return response.data;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roles'] });
    },
  });
}

// System logs hooks
export function useLogs(filters?: LogFilters) {
  return useQuery({
    queryKey: ['logs', filters],
    queryFn: async () => {
      const { data } = await api.get<SystemLog[]>('/logs', { params: filters });
      return data;
    },
  });
}