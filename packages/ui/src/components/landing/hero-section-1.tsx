'use client';

import { CircleDashed, Inbox } from 'lucide-react';

import { Button } from '../ui/button';
import { Input } from '../ui/input';

export interface LandingHeroSection1Props {
  onStartNow?: () => void;
  onHugMe?: () => void;
}

export function LandingHeroSection1({
  onStartNow,
  onHugMe,
}: LandingHeroSection1Props) {
  return (
    <section className="bg-background flex min-h-screen w-full justify-center px-4 py-12">
      <div className="flex w-full max-w-360 flex-row items-center gap-8">
        {/* Left content */}
        <div className="flex flex-1 flex-col justify-center gap-6">
          <div className="flex flex-col gap-3">
            <h1 className="text-foreground text-[48px] leading-tight font-semibold tracking-[-0.03em]">
              Connect every idea
            </h1>
            <p className="text-muted-foreground max-w-md text-base leading-relaxed font-normal">
              Turn scattered notes into a shared universe. Visualize your
              team&apos;s knowledge graph and collaborate in real-time with
              Notopia.
            </p>
          </div>

          {/* Contact pill input */}
          <div className="border-input bg-background relative flex w-full max-w-85 items-center gap-2 rounded-full border px-4 py-3 shadow-sm">
            <CircleDashed className="text-muted-foreground h-5 w-5 shrink-0" />
            <Input
              placeholder="Contact us"
              className="text-foreground placeholder:text-muted-foreground h-auto border-0 bg-transparent p-0 text-sm shadow-none focus-visible:ring-0"
            />
            <Inbox className="text-muted-foreground h-5 w-5 shrink-0" />
          </div>

          {/* CTA buttons */}
          <div className="flex items-center gap-4">
            <Button
              className="bg-primary text-primary-foreground hover:bg-primary/90 h-auto rounded-lg px-6 py-3 text-sm font-semibold"
              onClick={onStartNow}
            >
              Start Now
            </Button>
            <Button
              className="bg-secondary text-secondary-foreground hover:bg-secondary/90 h-auto rounded-lg border-0 px-6 py-3 text-sm font-semibold shadow-none"
              onClick={onHugMe}
            >
              Hug Me
            </Button>
          </div>
        </div>

        {/* Right content */}
        <div className="from-muted to-muted/50 flex flex-1 flex-col items-center justify-center gap-6 rounded-2xl bg-linear-to-br p-10">
          <div className="flex flex-col gap-4 text-center">
            <h2 className="text-foreground text-3xl font-semibold">
              Visualize Your Knowledge
            </h2>
            <p className="text-muted-foreground text-base leading-relaxed">
              Notopia transforms the way teams organize, visualize, and
              collaborate on shared information. Create powerful knowledge
              graphs and unlock new insights.
            </p>
          </div>
          <div className="mt-4 grid w-full grid-cols-2 gap-4">
            <div className="border-border bg-card rounded-lg border p-5 transition-shadow hover:shadow-md">
              <p className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                Real-time
              </p>
              <p className="text-foreground text-sm font-medium">
                Instant Sync
              </p>
            </div>
            <div className="border-border bg-card rounded-lg border p-5 transition-shadow hover:shadow-md">
              <p className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                Collaborative
              </p>
              <p className="text-foreground text-sm font-medium">
                Team Workflow
              </p>
            </div>
            <div className="border-border bg-card rounded-lg border p-5 transition-shadow hover:shadow-md">
              <p className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                Intelligent
              </p>
              <p className="text-foreground text-sm font-medium">
                Knowledge Graph
              </p>
            </div>
            <div className="border-border bg-card rounded-lg border p-5 transition-shadow hover:shadow-md">
              <p className="text-muted-foreground mb-2 text-xs font-semibold tracking-wide uppercase">
                Secure
              </p>
              <p className="text-foreground text-sm font-medium">
                Privacy First
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
