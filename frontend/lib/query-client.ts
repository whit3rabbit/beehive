import { QueryClient } from '@tanstack/react-query';
import { Socket, io } from 'socket.io-client';

// Create WebSocket connection
const socket: Socket = io('http://localhost:8080/ws');

// Configure QueryClient
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5000, // Consider data fresh for 5 seconds
      gcTime: 300000,  // Keep inactive data in cache for 5 minutes
    },
  },
});

// Set up socket event listeners
socket.on('connect', () => {
  console.log('WebSocket connected');
});

socket.on('disconnect', () => {
  console.log('WebSocket disconnected');
});

// Invalidate queries based on WebSocket events
socket.on('agent:update', () => {
  queryClient.invalidateQueries({ queryKey: ['agents'] });
});

socket.on('task:update', () => {
  queryClient.invalidateQueries({ queryKey: ['tasks'] });
});

// Export socket for use in components if needed
export { socket };