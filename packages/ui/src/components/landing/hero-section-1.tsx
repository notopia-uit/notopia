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
    <section className="flex w-full justify-center bg-white px-4 py-12">
      <div className="flex w-full max-w-[1440px] flex-row items-stretch gap-8">
        {/* Left content */}
        <div className="flex flex-1 flex-col justify-start gap-6">
          <div className="flex flex-col gap-3">
            <h1 className="text-[48px] leading-tight font-semibold tracking-[-0.03em] text-black">
              Connect every idea
            </h1>
            <p className="max-w-md text-base leading-relaxed font-normal text-slate-700">
              Turn scattered notes into a shared universe. Visualize your
              team&apos;s knowledge graph and collaborate in real-time with
              Notopia.
            </p>
          </div>

          {/* Contact pill input */}
          <div className="relative flex w-full max-w-[340px] items-center gap-2 rounded-full border border-slate-200 bg-white px-4 py-3 shadow-sm">
            <CircleDashed className="h-5 w-5 shrink-0 text-slate-500" />
            <Input
              placeholder="Contact us"
              className="h-auto border-0 bg-transparent p-0 text-sm text-slate-700 shadow-none placeholder:text-slate-400 focus-visible:ring-0"
            />
            <Inbox className="h-5 w-5 shrink-0 text-slate-500" />
          </div>

          {/* CTA buttons */}
          <div className="flex items-center gap-4">
            <Button
              className="h-auto rounded-lg bg-slate-900 px-6 py-3 text-sm font-semibold text-white hover:bg-slate-800"
              onClick={onStartNow}
            >
              Start Now
            </Button>
            <Button
              className="h-auto rounded-lg border-0 bg-slate-100 px-6 py-3 text-sm font-semibold text-slate-900 shadow-none hover:bg-slate-200"
              onClick={onHugMe}
            >
              Hug Me
            </Button>
          </div>
        </div>

        {/* Right content */}
        <div className="flex flex-1 flex-col items-center justify-center gap-6 rounded-2xl bg-gradient-to-br from-slate-50 to-slate-100 p-10">
          <div className="flex flex-col gap-4 text-center">
            <h2 className="text-3xl font-semibold text-slate-900">
              Visualize Your Knowledge
            </h2>
            <p className="text-base leading-relaxed text-slate-600">
              Notopia transforms the way teams organize, visualize, and
              collaborate on shared information. Create powerful knowledge
              graphs and unlock new insights.
            </p>
          </div>
          <div className="mt-4 grid w-full grid-cols-2 gap-4">
            <div className="rounded-lg border border-slate-200 bg-white p-5 transition-shadow hover:shadow-md">
              <p className="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
                Real-time
              </p>
              <p className="text-sm font-medium text-slate-900">Instant Sync</p>
            </div>
            <div className="rounded-lg border border-slate-200 bg-white p-5 transition-shadow hover:shadow-md">
              <p className="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
                Collaborative
              </p>
              <p className="text-sm font-medium text-slate-900">
                Team Workflow
              </p>
            </div>
            <div className="rounded-lg border border-slate-200 bg-white p-5 transition-shadow hover:shadow-md">
              <p className="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
                Intelligent
              </p>
              <p className="text-sm font-medium text-slate-900">
                Knowledge Graph
              </p>
            </div>
            <div className="rounded-lg border border-slate-200 bg-white p-5 transition-shadow hover:shadow-md">
              <p className="mb-2 text-xs font-semibold tracking-wide text-slate-500 uppercase">
                Secure
              </p>
              <p className="text-sm font-medium text-slate-900">
                Privacy First
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
