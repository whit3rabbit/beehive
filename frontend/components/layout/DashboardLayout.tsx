'use client';

import React from 'react';
import { SideNav } from './SideNav';
import { TopBar } from './TopBar';
import { BreadcrumbNav } from './BreadcrumbNav';

interface DashboardLayoutProps {
  children: React.ReactNode;
}

export const DashboardLayout: React.FC<DashboardLayoutProps> = ({ children }) => {
  return (
    <div className="min-h-screen bg-background">
      <TopBar />
      <div className="flex h-screen pt-16">
        <SideNav />
        <main className="flex-1 overflow-y-auto p-6 bg-background">
          <BreadcrumbNav />
          {children}
        </main>
      </div>
    </div>
  );
};