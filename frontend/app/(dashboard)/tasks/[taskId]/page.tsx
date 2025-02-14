'use client';

import { useTask } from '@/hooks/use-api';
import { TaskDetails } from '@/components/tasks/task-details';
import { isValidObjectId } from '@/lib/utils';
import { notFound } from 'next/navigation';
import { Loader2 } from 'lucide-react';

interface TaskDetailPageProps {
  params: {
    taskId: string;
  };
}

export default function TaskDetailPage({ params }: TaskDetailPageProps) {
  if (!isValidObjectId(params.taskId)) {
    notFound();
  }

  const { data: task, isLoading, error } = useTask(params.taskId);

  if (isLoading) {
    return (
      <div className="flex h-[calc(100vh-4rem)] items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !task) {
    return (
      <div className="flex h-[calc(100vh-4rem)] items-center justify-center">
        <div className="text-center">
          <h2 className="text-xl font-semibold">Task Not Found</h2>
          <p className="text-muted-foreground">
            The task you're looking for doesn't exist or you don't have permission to view it.
          </p>
        </div>
      </div>
    );
  }

  return <TaskDetails task={task} />;
}