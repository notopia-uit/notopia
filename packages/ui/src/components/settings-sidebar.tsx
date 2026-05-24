'use client';

import { cn } from '@notopia-uit/ui/lib/shadcn/utils';
import { CreditCard, Plug, Settings, Shield, Users } from 'lucide-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

interface SettingsSidebarProps {
  workspaceId: string;
}

export function SettingsSidebar({ workspaceId }: SettingsSidebarProps) {
  const pathname = usePathname();

  //TODO: This can be moved to a separate file if we want to reuse it in the future
  const navItems = [
    {
      name: 'General',
      href: `/workspace/${workspaceId}/settings/general`,
      icon: Settings,
    },
    {
      name: 'Members',
      href: `/workspace/${workspaceId}/settings/members`,
      icon: Users,
    },
    {
      name: 'Billing',
      href: `/workspace/${workspaceId}/settings/billing`,
      icon: CreditCard,
    },
    {
      name: 'Integrations',
      href: `/workspace/${workspaceId}/settings/integrations`,
      icon: Plug,
    },
    {
      name: 'Advanced',
      href: `/workspace/${workspaceId}/settings/advanced`,
      icon: Shield,
    },
  ];

  return (
    <nav className="flex space-x-2 lg:flex-col lg:space-y-1 lg:space-x-0">
      {navItems.map((item) => {
        const isActive = pathname === item.href;
        const Icon = item.icon;

        return (
          <Link
            key={item.name}
            href={item.href}
            className={cn(
              `flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors`,
              isActive
                ? 'bg-accent text-accent-foreground'
                : `text-muted-foreground hover:bg-secondary hover:text-foreground`
            )}
          >
            <Icon className="size-4" />
            {item.name}
          </Link>
        );
      })}
    </nav>
  );
}
