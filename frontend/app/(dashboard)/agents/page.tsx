'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Plus } from 'lucide-react';
import { useAgents } from '@/hooks/use-api';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/agents/data-table';

export default function AgentsPage() {
  const { data: agents, isLoading } = useAgents();

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-semibold">Agents</h1>
        <Link href="/agents/create">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Add Agent
          </Button>
        </Link>
      </div>
      
      <DataTable data={agents ?? []} isLoading={isLoading} />
    </div>
  );
}