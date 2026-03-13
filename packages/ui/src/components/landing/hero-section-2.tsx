'use client';

import { Inbox } from 'lucide-react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';

export interface LandingHeroSection2Props {
  onStartNow?: () => void;
  onHugMe?: () => void;
}

export function LandingHeroSection2({
  onStartNow,
  onHugMe,
}: LandingHeroSection2Props) {
  return (
    <section className="from-background via-muted/30 to-background w-full bg-gradient-to-b">
      <div className="mx-auto flex min-h-125 max-w-360 flex-col items-center justify-center gap-8 px-6 py-20">
        {/* Contact pill input */}
        <div className="border-border bg-background relative flex w-full max-w-85 items-center gap-2 rounded-full border px-4 py-3 shadow-sm transition-shadow hover:shadow-md">
          <Input
            placeholder="Contact us"
            className="text-foreground placeholder:text-muted-foreground h-auto border-0 bg-transparent p-0 text-sm shadow-none focus-visible:ring-0"
          />
          <Inbox className="text-muted-foreground h-5 w-5 shrink-0" />
        </div>

        {/* Heading */}
        <h1 className="text-foreground max-w-4xl text-center text-[56px] leading-tight font-bold tracking-[-0.02em]">
          Stop letting your best ideas get lost in the noise
        </h1>

        {/* Subtext */}
        <p className="text-muted-foreground max-w-2xl text-center text-lg leading-relaxed font-normal">
          Turn scattered notes into a shared universe. Visualize your
          team&apos;s knowledge graph and collaborate in real-time with Notopia.
        </p>

        {/* CTA buttons */}
        <div className="flex items-center gap-4 pt-4">
          <Button
            className="bg-primary text-primary-foreground hover:bg-primary/90 h-auto rounded-lg px-8 py-3 text-base font-semibold transition-colors"
            onClick={onStartNow}
          >
            Start Now
          </Button>
          <Button
            className="bg-secondary text-secondary-foreground hover:bg-secondary/90 h-auto rounded-lg border-0 px-8 py-3 text-base font-semibold shadow-none transition-colors"
            onClick={onHugMe}
          >
            Hug Me
          </Button>
        </div>
      </div>
    </section>
  );
}
