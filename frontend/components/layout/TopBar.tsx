'use client';

import { Bell, User, LogOut } from 'lucide-react';
import { ModeToggle } from '@/components/mode-toggle';
import { Button } from "@/components/ui/button"

export const TopBar: React.FC = () => {
  return (
    <header className="fixed top-0 z-40 w-full bg-background border-b">
      <div className="flex h-16 items-center justify-between px-4">
        <div className="flex items-center">
          <h1 className="text-xl font-semibold text-foreground">Beehive Admin</h1>
        </div>
        
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="icon" className="relative">
            <Bell className="h-5 w-5" />
          </Button>
          
          <ModeToggle />
          
          <div className="flex items-center gap-2">
            <User className="h-8 w-8" />
            <span className="text-sm font-medium text-foreground">Admin User</span>
          </div>
          
          <Button variant="ghost" size="icon">
            <LogOut className="h-5 w-5" />
          </Button>
        </div>
      </div>
    </header>
  );
};