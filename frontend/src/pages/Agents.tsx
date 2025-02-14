// pages/Agents.tsx
import { AgentList } from '@/components/agents/AgentList';
import { DashboardLayout } from '@/layouts/DashboardLayout';

export const AgentsPage = () => {
  return (
    <DashboardLayout>
      <h1>Agents</h1>
      <AgentList />
    </DashboardLayout>
  );
};
