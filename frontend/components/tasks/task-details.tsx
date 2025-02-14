'use client';

import { useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { TaskOutput } from '@/components/tasks/task-output';
import { formatTimestamp } from '@/lib/utils';
import type { Task } from '@/types/mongodb';

interface TaskDetailsProps {
  task: Task;
}

export function TaskDetails({ task }: TaskDetailsProps) {
  const router = useRouter();

  return (
    <div className="space-y-6">
      <div className="flex items-center space-x-4">
        <Button
          variant="ghost"
          size="sm"
          onClick={() => router.back()}
        >
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <h1 className="text-2xl font-semibold">Task Details</h1>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Task Information</CardTitle>
            <CardDescription>Basic task details</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Status</p>
              <p className="mt-1">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium
                  ${task.status === 'completed' ? 'bg-green-100 text-green-800' :
                    task.status === 'running' ? 'bg-blue-100 text-blue-800' :
                    task.status === 'failed' ? 'bg-red-100 text-red-800' :
                    'bg-gray-100 text-gray-800'}`
                }>
                  {task.status}
                </span>
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Type</p>
              <p className="mt-1">{task.type}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Created</p>
              <p className="mt-1">{formatTimestamp(task.created_at)}</p>
            </div>
            {task.started_at && (
              <div>
                <p className="text-sm font-medium text-muted-foreground">Started</p>
                <p className="mt-1">{formatTimestamp(task.started_at)}</p>
              </div>
            )}
            <div>
              <p className="text-sm font-medium text-muted-foreground">Parameters</p>
              <pre className="mt-1 p-2 bg-muted rounded-md text-xs overflow-auto">
                {JSON.stringify(task.parameters, null, 2)}
              </pre>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Task Output</CardTitle>
            <CardDescription>Execution results and logs</CardDescription>
          </CardHeader>
          <CardContent>
            <TaskOutput task={task} />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}