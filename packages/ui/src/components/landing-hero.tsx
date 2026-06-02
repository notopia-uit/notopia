'use client';

import Link from 'next/link';
import { CircleDashed, Inbox } from 'lucide-react';

import { Button } from './shadcn/button';
import { Card, CardContent } from './shadcn/card';
import { Input } from './shadcn/input';

const FEATURE_TILES = [
  { label: 'Real-time', title: 'Instant Sync' },
  { label: 'Collaborative', title: 'Team Workflow' },
  { label: 'Intelligent', title: 'Knowledge Graph' },
  { label: 'Secure', title: 'Privacy First' },
];

export default function LandingHero() {
  return (
    <section className="bg-background flex min-h-screen w-full justify-center px-6 py-12">
      <div className="flex w-full max-w-6xl flex-col items-center gap-12 lg:flex-row lg:items-center">
        <div className="flex flex-1 flex-col justify-center gap-6">
          <div className="flex flex-col gap-3">
            <h1 className="text-foreground text-4xl/tight  font-semibold tracking-tight sm:text-5xl lg:text-[52px]">
              Connect every idea
            </h1>
            <p className="text-muted-foreground max-w-md text-base/relaxed ">
              Turn scattered notes into a shared universe. Visualize your team&apos;s knowledge
              graph and collaborate in real-time with Notopia.
            </p>
          </div>

          <div className="border-input bg-background flex w-full max-w-sm items-center gap-2 rounded-full border px-4 py-3 shadow-sm">
            <CircleDashed className="text-muted-foreground size-5  shrink-0" />
            <Input
              placeholder="Contact us"
              className="text-foreground placeholder:text-muted-foreground h-auto border-0 bg-transparent p-0 text-sm shadow-none focus-visible:ring-0"
            />
            <Inbox className="text-muted-foreground size-5  shrink-0" />
          </div>

          <div className="flex items-center gap-3">
            <Button asChild size="lg" className="rounded-lg px-6">
              <Link href="/workspace">Get Started</Link>
            </Button>
            <Button asChild size="lg" variant="secondary" className="rounded-lg px-6">
              <Link href="#features">Learn More</Link>
            </Button>
          </div>
        </div>

        <div className="from-muted to-muted/40 flex flex-1 flex-col items-center justify-center gap-6 rounded-2xl bg-linear-to-br p-10">
          <div className="flex flex-col gap-3 text-center">
            <h2 className="text-foreground text-2xl font-semibold">Visualize Your Knowledge</h2>
            <p className="text-muted-foreground text-sm/relaxed ">
              Notopia transforms the way teams organize, visualize, and collaborate on shared
              information. Create powerful knowledge graphs and unlock new insights.
            </p>
          </div>

          <div className="mt-2 grid w-full grid-cols-2 gap-3">
            {FEATURE_TILES.map((tile) => (
              <Card
                key={tile.label}
                className="border-border bg-card transition-shadow hover:shadow-md"
              >
                <CardContent className="p-5">
                  <p className="text-muted-foreground mb-1 text-xs font-semibold tracking-wide uppercase">
                    {tile.label}
                  </p>
                  <p className="text-foreground text-sm font-medium">{tile.title}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
