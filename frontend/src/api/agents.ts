import axios from 'axios';
import { Agent } from './types';

export const getAgents = async (): Promise<Agent[]> => {
  const response = await axios.get('/api/agents');
  return response.data;
};
