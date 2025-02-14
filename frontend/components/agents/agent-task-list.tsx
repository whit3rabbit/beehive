'use client';

import { formatTimestamp } from '@/lib/utils';
import type { Task } from '@/types/mongodb';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

interface AgentTaskListProps {
  tasks: Task[];
  isLoading: boolean;
}

export function AgentTaskList({ tasks, isLoading }: AgentTaskListProps) {
  if (isLoading) {
    return <div>Loading tasks...</div>;
  }

  if (tasks.length === 0) {
    return <div className="text-sm text-muted-foreground">No tasks found</div>;
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Type</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Created</TableHead>
          <TableHead>Last Updated</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {tasks.map((task) => (
          <TableRow key={task._id}>
            <TableCell className="font-medium">{task.type}</TableCell>
            <TableCell>
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                ${task.status === 'completed' ? 'bg-green-100 text-green-800' :
                  task.status === 'running' ? 'bg-blue-100 text-blue-800' :
                  task.status === 'failed' ? 'bg-red-100 text-red-800' :
                  'bg-gray-100 text-gray-800'}`
              }>
                {task.status}
              </span>
            </TableCell>
            <TableCell>{formatTimestamp(task.created_at)}</TableCell>
            <TableCell>{formatTimestamp(task.updated_at || task.created_at)}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}