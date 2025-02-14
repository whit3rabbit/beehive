'use client';

import { ChevronRight } from 'lucide-react';
import { usePathname } from 'next/navigation';
import Link from 'next/link';

export const BreadcrumbNav: React.FC = () => {
  const pathname = usePathname();
  const segments = pathname.split('/').filter(Boolean);
  
  return (
    <nav className="mb-4">
      <ol className="flex items-center space-x-2">
        <li>
          <Link href="/dashboard" className="text-muted-foreground hover:text-foreground">
            Home
          </Link>
        </li>
        
        {segments.map((segment, index) => {
          const path = `/${segments.slice(0, index + 1).join('/')}`;
          const isLast = index === segments.length - 1;
          
          return (
            <li key={path} className="flex items-center">
              <ChevronRight className="h-4 w-4 text-muted-foreground" />
              <Link
                href={path}
                className={`ml-2 capitalize ${
                  isLast ? 'text-foreground font-medium' : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                {segment}
              </Link>
            </li>
          );
        })}
      </ol>
    </nav>
  );
};