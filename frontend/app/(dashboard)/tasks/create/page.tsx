'use client';

import { TaskForm } from '@/components/tasks/task-form';

export default function CreateTaskPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Create New Task</h1>
      <p className="text-muted-foreground">
        Create a new task to be executed by an agent.
      </p>
      <TaskForm />
    </div>
  );
}