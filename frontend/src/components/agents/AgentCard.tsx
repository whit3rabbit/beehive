import { Agent } from '@/api/types';

interface AgentCardProps {
  agent: Agent;
}

export const AgentCard: React.FC<AgentCardProps> = ({ agent }) => {
  return (
    <div>
      <h3>{agent.nickname || agent.hostname}</h3>
      <p>Hostname: {agent.hostname}</p>
      <p>UUID: {agent.uuid}</p>
      <p>Status: {agent.status}</p>
    </div>
  );
};
