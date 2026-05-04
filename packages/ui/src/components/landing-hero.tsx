import Link from 'next/link';

import { cn } from '../lib/shadcn/utils';
import { buttonVariants } from './shadcn/button';

export default function LandingHero() {
  return (
    <section className="space-y-6 pt-6 pb-8 md:pt-10 md:pb-12 lg:py-32">
      <div className="container flex max-w-5xl flex-col items-center gap-4 text-center">
        <h1 className="font-heading text-3xl sm:text-5xl md:text-6xl lg:text-7xl">
          An example app built using Next.js 13 server components.
        </h1>
        <p className="text-muted-foreground max-w-2xl leading-normal sm:text-xl/8">
          I&apos;m building a web app with Next.js 13 and open sourcing everything. Follow along as
          we figure this out together.
        </p>
        <div className="space-x-4">
          <Link href="/login" className={cn(buttonVariants({ size: 'lg' }))}>
            Get Started
          </Link>
          <Link
            href="/"
            target="_blank"
            rel="noreferrer"
            className={cn(buttonVariants({ variant: 'outline', size: 'lg' }))}
          >
            GitHub
          </Link>
        </div>
      </div>
    </section>
  );
}
