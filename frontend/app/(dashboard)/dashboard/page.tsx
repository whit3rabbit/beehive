'use client';

import { useAgents, useTasks } from "@/hooks/use-api";

export default function DashboardPage() {
  const { data: agents } = useAgents();
  const { data: tasks } = useTasks();

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Dashboard</h1>
      
      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-card text-card-foreground p-6 rounded-lg border shadow">
          <h3 className="text-sm font-medium text-muted-foreground">Total Agents</h3>
          <p className="text-3xl font-bold mt-2">{agents?.length ?? 0}</p>
        </div>
        <div className="bg-card text-card-foreground p-6 rounded-lg border shadow">
          <h3 className="text-sm font-medium text-muted-foreground">Active Tasks</h3>
          <p className="text-3xl font-bold mt-2">
            {tasks?.filter(task => task.status === 'running').length ?? 0}
          </p>
        </div>
        <div className="bg-card text-card-foreground p-6 rounded-lg border shadow">
          <h3 className="text-sm font-medium text-muted-foreground">System Health</h3>
          <p className="text-3xl font-bold mt-2 text-emerald-500">98%</p>
        </div>
      </div>
      
      {/* Recent Activity */}
      <div className="bg-card text-card-foreground p-6 rounded-lg border shadow">
        <h2 className="text-xl font-medium mb-4">Recent Activity</h2>
        <div className="space-y-4">
          {tasks && tasks.length > 0 ? (
            tasks.slice(0, 5).map(task => (
              <div key={task._id} className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">
                  {task.type} - {task.status}
                </span>
                <span className="text-sm">
                  {new Date(task.created_at).toLocaleString()}
                </span>
              </div>
            ))
          ) : (
            <div className="text-sm text-muted-foreground">
              No recent activity to display
            </div>
          )}
        </div>
      </div>
    </div>
  );
}