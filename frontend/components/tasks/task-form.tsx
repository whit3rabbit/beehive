'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useCreateTask, useAgents } from '@/hooks/use-api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import type { TaskType } from '@/types/mongodb';

const TASK_TYPES: { value: TaskType; label: string }[] = [
  { value: 'command_shell', label: 'Command Shell' },
  { value: 'file_operation', label: 'File Operation' },
  { value: 'ui_automation', label: 'UI Automation' },
  { value: 'browser_automation', label: 'Browser Automation' },
];

export function TaskForm() {
  const router = useRouter();
  const [error, setError] = useState<string>('');
  const [taskType, setTaskType] = useState<TaskType>('command_shell');
  const createTask = useCreateTask();
  const { data: agents } = useAgents();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');

    const formData = new FormData(e.currentTarget);
    const agent_id = formData.get('agent_id') as string;
    const timeout = parseInt(formData.get('timeout') as string) || 3600;

    // Get parameters based on task type
    const parameters: Record<string, unknown> = {};
    switch (taskType) {
      case 'command_shell':
        parameters.command = formData.get('command');
        break;
      case 'file_operation':
        parameters.operation = formData.get('operation');
        parameters.path = formData.get('path');
        break;
      case 'ui_automation':
      case 'browser_automation':
        parameters.script = formData.get('script');
        break;
    }

    try {
      const task = await createTask.mutateAsync({
        agent_id,
        type: taskType,
        parameters,
        timeout,
      });
      router.push(`/tasks/${task._id}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create task');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Basic Information</CardTitle>
          <CardDescription>General task settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="agent">Agent</Label>
            <Select name="agent_id" required>
              <SelectTrigger>
                <SelectValue placeholder="Select an agent" />
              </SelectTrigger>
              <SelectContent>
                {agents?.map((agent) => (
                  <SelectItem key={agent._id} value={agent._id}>
                    {agent.nickname || agent.hostname}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="type">Task Type</Label>
            <Select
              name="type"
              value={taskType}
              onValueChange={(value) => setTaskType(value as TaskType)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select task type" />
              </SelectTrigger>
              <SelectContent>
                {TASK_TYPES.map((type) => (
                  <SelectItem key={type.value} value={type.value}>
                    {type.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="timeout">Timeout (seconds)</Label>
            <Input
              id="timeout"
              name="timeout"
              type="number"
              defaultValue="3600"
              min="0"
            />
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Task Parameters</CardTitle>
          <CardDescription>Configure task-specific settings</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {taskType === 'command_shell' && (
            <div className="space-y-2">
              <Label htmlFor="command">Command</Label>
              <Textarea
                id="command"
                name="command"
                placeholder="Enter command to execute"
                required
              />
            </div>
          )}

          {taskType === 'file_operation' && (
            <>
              <div className="space-y-2">
                <Label htmlFor="operation">Operation</Label>
                <Select name="operation" required>
                  <SelectTrigger>
                    <SelectValue placeholder="Select operation" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="read">Read</SelectItem>
                    <SelectItem value="write">Write</SelectItem>
                    <SelectItem value="delete">Delete</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="path">File Path</Label>
                <Input
                  id="path"
                  name="path"
                  placeholder="Enter file path"
                  required
                />
              </div>
            </>
          )}

          {(taskType === 'ui_automation' || taskType === 'browser_automation') && (
            <div className="space-y-2">
              <Label htmlFor="script">Automation Script</Label>
              <Textarea
                id="script"
                name="script"
                placeholder="Enter automation script"
                required
                className="font-mono"
                rows={10}
              />
            </div>
          )}
        </CardContent>
      </Card>

      <div className="flex justify-end space-x-4">
        <Button
          type="button"
          variant="outline"
          onClick={() => router.back()}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={createTask.isPending}>
          {createTask.isPending ? 'Creating...' : 'Create Task'}
        </Button>
      </div>
    </form>
  );
}