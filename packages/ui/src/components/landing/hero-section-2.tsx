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
    <section className="w-full bg-linear-to-b from-white via-slate-50 to-white">
      <div className="mx-auto flex min-h-125 max-w-360 flex-col items-center justify-center gap-8 px-6 py-20">
        {/* Contact pill input */}
        <div className="relative flex w-full max-w-85 items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-3 shadow-sm transition-shadow hover:shadow-md">
          <Input
            placeholder="Contact us"
            className="h-auto border-0 bg-transparent p-0 text-sm text-slate-700 shadow-none placeholder:text-slate-400 focus-visible:ring-0"
          />
          <Inbox className="h-5 w-5 shrink-0 text-slate-500" />
        </div>

        {/* Heading */}
        <h1 className="max-w-4xl text-center text-[56px] leading-tight font-bold tracking-[-0.02em] text-slate-900">
          Stop letting your best ideas get lost in the noise
        </h1>

        {/* Subtext */}
        <p className="max-w-2xl text-center text-lg leading-relaxed font-normal text-slate-600">
          Turn scattered notes into a shared universe. Visualize your
          team&apos;s knowledge graph and collaborate in real-time with Notopia.
        </p>

        {/* CTA buttons */}
        <div className="flex items-center gap-4 pt-4">
          <Button
            className="h-auto rounded-lg bg-slate-900 px-8 py-3 text-base font-semibold text-white transition-colors hover:bg-slate-800"
            onClick={onStartNow}
          >
            Start Now
          </Button>
          <Button
            className="h-auto rounded-lg border-0 bg-slate-100 px-8 py-3 text-base font-semibold text-slate-900 shadow-none transition-colors hover:bg-slate-200"
            onClick={onHugMe}
          >
            Hug Me
          </Button>
        </div>
      </div>
    </section>
  );
}
