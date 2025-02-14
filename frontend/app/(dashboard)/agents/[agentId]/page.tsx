'use client';

import { useAgent } from '@/hooks/use-api';
import { AgentDetails } from '@/components/agents/agent-details';
import { isValidObjectId } from '@/lib/utils';
import { notFound } from 'next/navigation';

interface AgentDetailPageProps {
  params: {
    agentId: string;
  };
}

export default function AgentDetailPage({ params }: AgentDetailPageProps) {
  // Validate MongoDB ObjectId format
  if (!isValidObjectId(params.agentId)) {
    notFound();
  }

  const { data: agent, isLoading, error } = useAgent(params.agentId);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error || !agent) {
    return <div>Agent not found</div>;
  }

  return <AgentDetails agent={agent} />;
}