'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useCreateAgent, useUpdateAgent, useRoles } from '@/hooks/use-api';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { Agent } from '@/types/mongodb';

interface AgentFormProps {
  mode: 'create' | 'edit';
  agent?: Agent;
}

export function AgentForm({ mode, agent }: AgentFormProps) {
  const router = useRouter();
  const [error, setError] = useState<string>('');
  const createAgent = useCreateAgent();
  const updateAgent = useUpdateAgent();
  const { data: roles } = useRoles();

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');

    const formData = new FormData(e.currentTarget);
    const data = {
      hostname: formData.get('hostname') as string,
      nickname: formData.get('nickname') as string,
      role: formData.get('role') as string,
      mac_hash: mode === 'create' ? formData.get('mac_hash') as string : undefined,
    };

    try {
      if (mode === 'create') {
        const newAgent = await createAgent.mutateAsync(data);
        router.push(`/agents/${newAgent._id}`);
      } else if (agent?._id) {
        await updateAgent.mutateAsync({
          id: agent._id,
          data: {
            nickname: data.nickname,
            role: data.role,
          },
        });
        router.push(`/agents/${agent._id}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to process agent');
    }
  };

  const isLoading = mode === 'create' ? createAgent.isPending : updateAgent.isPending;

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="hostname">Hostname</Label>
          <Input
            id="hostname"
            name="hostname"
            defaultValue={agent?.hostname}
            required
            disabled={mode === 'edit'}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="nickname">Nickname (Optional)</Label>
          <Input
            id="nickname"
            name="nickname"
            defaultValue={agent?.nickname}
            placeholder="Friendly name for the agent"
          />
        </div>

        {mode === 'create' && (
          <div className="space-y-2">
            <Label htmlFor="mac_hash">MAC Address Hash</Label>
            <Input
              id="mac_hash"
              name="mac_hash"
              required
              placeholder="Unique MAC address hash"
            />
          </div>
        )}

        <div className="space-y-2">
          <Label htmlFor="role">Role</Label>
          <Select name="role" defaultValue={agent?.role}>
            <SelectTrigger>
              <SelectValue placeholder="Select a role" />
            </SelectTrigger>
            <SelectContent>
              {roles?.map((role) => (
                <SelectItem key={role._id} value={role._id}>
                  {role.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="flex justify-end space-x-4">
        <Button
          type="button"
          variant="outline"
          onClick={() => router.back()}
        >
          Cancel
        </Button>
        <Button type="submit" disabled={isLoading}>
          {isLoading ? (
            mode === 'create' ? 'Creating...' : 'Saving...'
          ) : (
            mode === 'create' ? 'Create Agent' : 'Save Changes'
          )}
        </Button>
      </div>
    </form>
  );
}