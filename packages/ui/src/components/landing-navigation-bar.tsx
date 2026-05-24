'use client';

import { Icons } from '@notopia-uit/ui/components/icons';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@notopia-uit/ui/components/shadcn/navigation-menu';
import useIsMobile from '@notopia-uit/ui/hooks/use-is-mobile';
import Link from 'next/link';

function NavigationButtonGroup() {
  return (
    <div className="flex items-center justify-end space-x-2 p-2">
      <Button variant="outline" size="icon" aria-label="Github">
        <Icons.Github />
      </Button>
      <Button variant="outline" size="icon" aria-label="Facebook">
        <Icons.Facebook />
      </Button>
    </div>
  );
}
function NavigationBarMenu() {
  const isMobile = useIsMobile();
  return (
    <NavigationMenu viewport={!isMobile}>
      <NavigationMenuList className="flex-wrap">
        <NavigationMenuItem>
          <NavigationMenuTrigger>Features</NavigationMenuTrigger>
        </NavigationMenuItem>
        <NavigationMenuItem>
          <NavigationMenuTrigger>Pricing</NavigationMenuTrigger>
        </NavigationMenuItem>
        <NavigationMenuItem>
          <NavigationMenuTrigger>Resources</NavigationMenuTrigger>
        </NavigationMenuItem>
      </NavigationMenuList>
    </NavigationMenu>
  );
}

function NavigationBarLogo() {
  return (
    <div className="flex items-center space-x-4 p-2">
      <Link href="/" className="hidden items-center justify-start space-x-2 p-2 sm:flex">
        <Icons.Logo />
        <span className="hidden font-bold sm:inline-block">Notopia</span>
      </Link>
    </div>
  );
}

export default function LandingNavigationBar() {
  return (
    <>
      <NavigationBarLogo />
      <NavigationBarMenu />
      <NavigationButtonGroup />
    </>
  );
}
