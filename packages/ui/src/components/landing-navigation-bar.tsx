'use client';

import { Icons } from '@notopia-uit/ui/components/icons';
import { Button } from '@notopia-uit/ui/components/ui/button';
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuList,
  NavigationMenuTrigger,
} from '@notopia-uit/ui/components/ui/navigation-menu';
import useIsMobile from '@notopia-uit/ui/hooks/use-is-mobile';
import Link from 'next/link';

function NavigationButtonGroup() {
  return (
    <div className="flex justify-end items-center space-x-2 px-2 py-2">
      <Button variant="outline" size="icon" aria-label="Github">
        <Icons.Github />
      </Button>
      <Button variant="outline" size="icon" aria-label="Twitter">
        <Icons.Twitter />
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
    <div className="flex items-center space-x-4 px-2 py-2">
      <Link
        href="/"
        className="hidden justify-start items-center sm:flex space-x-2 px-2 py-2"
      >
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
