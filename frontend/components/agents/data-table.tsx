'use client';

import Link from 'next/link';
import { Eye } from 'lucide-react';
import { formatTimestamp } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import type { Agent } from '@/types/mongodb';

interface DataTableProps {
  data: Agent[];
  isLoading: boolean;
}

export function DataTable({ data, isLoading }: DataTableProps) {
  if (isLoading) {
    return <div>Loading agents...</div>;
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Status</TableHead>
            <TableHead>Hostname</TableHead>
            <TableHead>Nickname</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Last Seen</TableHead>
            <TableHead className="w-[100px]">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center">
                No agents found
              </TableCell>
            </TableRow>
          ) : (
            data.map((agent) => (
              <TableRow key={agent._id}>
                <TableCell>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                    ${agent.status === 'active' ? 'bg-green-100 text-green-800' :
                      agent.status === 'inactive' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-red-100 text-red-800'}`
                  }>
                    {agent.status}
                  </span>
                </TableCell>
                <TableCell className="font-medium">{agent.hostname}</TableCell>
                <TableCell>{agent.nickname || '-'}</TableCell>
                <TableCell>{agent.role || '-'}</TableCell>
                <TableCell>{formatTimestamp(agent.last_seen)}</TableCell>
                <TableCell>
                  <Link href={`/agents/${agent._id}`}>
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