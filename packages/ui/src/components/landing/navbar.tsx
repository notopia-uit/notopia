'use client';

import { SiGithub } from '@icons-pack/react-simple-icons';

import { Button } from '../ui/button';
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '../ui/navigation-menu';

const NAV_ITEMS = [
  { label: 'Features' },
  { label: 'Pricing' },
  { label: 'Resources' },
];

function NotopiaSvgLogo() {
  return (
    <svg
      width="36"
      height="36"
      viewBox="0 0 36 36"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <rect width="36" height="36" rx="8" fill="#0F172A" />
      <path d="M10 26V10h3l10 12V10h3v16h-3L13 14v12h-3z" fill="#F8FAFC" />
    </svg>
  );
}

export interface LandingNavbarProps {
  onFeatures?: () => void;
  onPricing?: () => void;
  onResources?: () => void;
}

export function LandingNavbar(_props: LandingNavbarProps) {
  return (
    <header className="flex w-full justify-center bg-transparent px-4 py-4">
      <nav
        className="flex w-full max-w-360 items-center justify-between gap-8"
        aria-label="Main navigation"
      >
        {/* Logo */}
        <div className="flex shrink-0 items-center gap-2">
          <NotopiaSvgLogo />
          <span className="text-foreground text-[24px] leading-none font-semibold tracking-tight">
            Notopia
          </span>
        </div>

        {/* Navigation menu */}
        <div className="flex flex-1 justify-center">
          <NavigationMenu viewport={false}>
            <NavigationMenuList className="gap-2">
              {NAV_ITEMS.map((item) => (
                <NavigationMenuItem key={item.label}>
                  <NavigationMenuTrigger className="text-foreground hover:bg-muted h-auto rounded-lg bg-transparent px-4 py-[7.5px] text-sm font-semibold">
                    {item.label}
                  </NavigationMenuTrigger>
                </NavigationMenuItem>
              ))}
            </NavigationMenuList>
          </NavigationMenu>
        </div>

        {/* Social links */}
        <div className="flex shrink-0 items-center justify-end gap-2">
          <Button
            variant="outline"
            size="icon"
            className="border-border rounded-lg"
            aria-label="GitHub"
          >
            <SiGithub className="text-foreground h-5 w-5" />
          </Button>
        </div>
      </nav>
    </header>
  );
}
