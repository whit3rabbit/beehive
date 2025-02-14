'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Plus } from 'lucide-react';
import { useTasks } from '@/hooks/use-api';
import { Button } from '@/components/ui/button';
import { DataTable } from '@/components/tasks/data-table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { TaskStatus } from '@/types/mongodb';

export default function TasksPage() {
  const [status, setStatus] = useState<TaskStatus | 'all'>('all');
  const { data: tasks, isLoading } = useTasks(
    status !== 'all' ? { status } : undefined
  );

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-semibold">Tasks</h1>
        <Link href="/tasks/create">
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            Create Task
          </Button>
        </Link>
      </div>

      <div className="flex items-center gap-4">
        <div className="w-[200px]">
          <Select value={status} onValueChange={(value) => setStatus(value as TaskStatus | 'all')}>
            <SelectTrigger>
              <SelectValue placeholder="Filter by status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Statuses</SelectItem>
              <SelectItem value="queued">Queued</SelectItem>
              <SelectItem value="running">Running</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="failed">Failed</SelectItem>
              <SelectItem value="cancelled">Cancelled</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <DataTable data={tasks ?? []} isLoading={isLoading} />
    </div>
  );
}