'use client';

import Link from 'next/link';
import { ArrowLeft, Edit } from 'lucide-react';
import { useRouter } from 'next/navigation';
import { useTasks } from '@/hooks/use-api';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { AgentTaskList } from '@/components/agents/agent-task-list';
import type { Agent } from '@/types/mongodb';

interface AgentDetailsProps {
  agent: Agent;
}

export function AgentDetails({ agent }: AgentDetailsProps) {
  const router = useRouter();
  const { data: tasks, isLoading: isLoadingTasks } = useTasks({ agent_id: agent._id });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.back()}
          >
            <ArrowLeft className="h-4 w-4 mr-2" />
            Back
          </Button>
          <h1 className="text-2xl font-semibold">{agent.nickname || agent.hostname}</h1>
        </div>
        <Link href={`/agents/${agent._id}/edit`}>
          <Button>
            <Edit className="h-4 w-4 mr-2" />
            Edit Agent
          </Button>
        </Link>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Agent Details</CardTitle>
            <CardDescription>Basic information about this agent</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Status</p>
              <p className="mt-1">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                  ${agent.status === 'active' ? 'bg-green-100 text-green-800' :
                    agent.status === 'inactive' ? 'bg-yellow-100 text-yellow-800' :
                    'bg-red-100 text-red-800'}`
                }>
                  {agent.status}
                </span>
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Hostname</p>
              <p className="mt-1 font-mono text-sm">{agent.hostname}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Role</p>
              <p className="mt-1">{agent.role || 'No role assigned'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Last Seen</p>
              <p className="mt-1 text-sm">
                {new Date(agent.last_seen).toLocaleString()}
              </p>
            </div>
          </CardContent>
        </Card>

        <div className="md:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>Recent Tasks</CardTitle>
              <CardDescription>Tasks assigned to this agent</CardDescription>
            </CardHeader>
            <CardContent>
              <AgentTaskList
                tasks={tasks ?? []}
                isLoading={isLoadingTasks}
              />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}