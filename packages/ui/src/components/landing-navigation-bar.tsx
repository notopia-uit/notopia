'use client';

import Link from 'next/link';

import { Icons } from './icons';
import { Button } from './shadcn/button';
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuList,
  NavigationMenuTrigger,
} from './shadcn/navigation-menu';

const NAV_ITEMS = ['Features', 'Pricing', 'Resources'];

export default function LandingNavigationBar() {
  return (
    <div className="flex w-full items-center justify-between gap-6">
      <Link href="/" className="flex shrink-0 items-center gap-2">
        <span className="text-foreground text-xl font-semibold tracking-tight">
          Notopia
        </span>
      </Link>

      <div className="flex flex-1 justify-center">
        <NavigationMenu viewport={false}>
          <NavigationMenuList className="gap-1">
            {NAV_ITEMS.map((item) => (
              <NavigationMenuItem key={item}>
                <NavigationMenuTrigger className="text-foreground hover:bg-muted h-auto rounded-lg bg-transparent px-4 py-2 text-sm font-medium">
                  {item}
                </NavigationMenuTrigger>
              </NavigationMenuItem>
            ))}
          </NavigationMenuList>
        </NavigationMenu>
      </div>

      <div className="flex shrink-0 items-center gap-2">
        <Button variant="outline" size="icon" className="rounded-lg" aria-label="GitHub">
          <Icons.Github className="size-4" />
        </Button>
      </div>
    </div>
  );
}
