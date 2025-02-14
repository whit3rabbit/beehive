'use client';

import { AgentForm } from '@/components/agents/agent-form';

export default function CreateAgentPage() {
  return (
    <div className="max-w-2xl mx-auto space-y-4">
      <h1 className="text-2xl font-semibold">Create New Agent</h1>
      <AgentForm mode="create" />
    </div>
  );
}