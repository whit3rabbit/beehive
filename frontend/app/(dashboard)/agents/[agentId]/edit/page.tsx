'use client';

import { useAgent } from '@/hooks/use-api';
import { AgentForm } from '@/components/agents/agent-form';
import { isValidObjectId } from '@/lib/utils';
import { notFound } from 'next/navigation';

interface EditAgentPageProps {
  params: {
    agentId: string;
  };
}

export default function EditAgentPage({ params }: EditAgentPageProps) {
  if (!isValidObjectId(params.agentId)) {
    notFound();
  }

  const { data: agent, isLoading } = useAgent(params.agentId);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!agent) {
    return <div>Agent not found</div>;
  }

  return (
    <div className="max-w-2xl mx-auto space-y-4">
      <h1 className="text-2xl font-semibold">Edit Agent</h1>
      <AgentForm mode="edit" agent={agent} />
    </div>
  );
}