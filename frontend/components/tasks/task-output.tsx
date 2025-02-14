'use client';

import { useEffect, useRef } from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';
import type { Task } from '@/types/mongodb';

interface TaskOutputProps {
  task: Task;
}

export function TaskOutput({ task }: TaskOutputProps) {
  const outputRef = useRef<HTMLPreElement>(null);

  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [task.output]);

  if (!task.output) {
    return (
      <div className="text-sm text-muted-foreground">
        No output available
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {task.output.error && (
        <Alert variant="destructive">
          <AlertDescription>{task.output.error}</AlertDescription>
        </Alert>
      )}
      
      {task.output.logs && (
        <pre
          ref={outputRef}
          className="p-4 bg-muted rounded-md text-xs font-mono whitespace-pre-wrap overflow-auto max-h-[400px]"
        >
          {task.output.logs}
        </pre>
      )}
    </div>
  );
}