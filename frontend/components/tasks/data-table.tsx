'use client';

import Link from 'next/link';
import { Eye } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { formatTimestamp } from '@/lib/utils';
import type { Task } from '@/types/mongodb';

interface DataTableProps {
  data: Task[];
  isLoading: boolean;
}

export function DataTable({ data, isLoading }: DataTableProps) {
  if (isLoading) {
    return <div>Loading tasks...</div>;
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>Type</TableHead>
            <TableHead>Agent</TableHead>
            <TableHead>Created</TableHead>
            <TableHead>Started</TableHead>
            <TableHead className="w-[100px]">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center">
                No tasks found
              </TableCell>
            </TableRow>
          ) : (
            data.map((task) => (
              <TableRow key={task._id}>
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
                <TableCell>{task.type}</TableCell>
                <TableCell>{task.agent_id}</TableCell>
                <TableCell>{formatTimestamp(task.created_at)}</TableCell>
                <TableCell>
                  {task.started_at ? formatTimestamp(task.started_at) : '-'}
                </TableCell>
                <TableCell>
                  <Link href={`/tasks/${task._id}`}>
                    <Button variant="ghost" size="sm">
                      <Eye className="h-4 w-4 mr-2" />
                      View
                    </Button>
                  </Link>
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}