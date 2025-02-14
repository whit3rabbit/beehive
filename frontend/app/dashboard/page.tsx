'use client';

import { DashboardLayout } from '@/components/layout/DashboardLayout';

export default function DashboardPage() {
  return (
    <DashboardLayout>
      <div className="space-y-6">
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        
        {/* Overview Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-card text-card-foreground p-6 rounded-lg border shadow">
            <h3 className="text-sm font-medium text-muted-foreground">Total Agents</h3>
            <p className="text-3xl font-bold mt-2">24</p>
          </div>
          <div className="bg-card text-card-foreground p-6 rounded-lg border shadow">
            <h3 className="text-sm font-medium text-muted-foreground">Active Tasks</h3>
            <p className="text-3xl font-bold mt-2">12</p>
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
            <div className="text-sm text-muted-foreground">
              No recent activity to display
            </div>
          </div>
        </div>
      </div>
    </DashboardLayout>
  );
}