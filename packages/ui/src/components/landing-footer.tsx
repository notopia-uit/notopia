import { Separator } from './shadcn/separator';
import { ModeToggle } from './theme-mode-toggle';

export default function LandingFooter() {
  return (
    <footer className="border-t">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 p-6 md:flex-row">
        <div>
          <div className="space-y-1">
            <h4 className="text-sm font-semibold leading-none">Notopia</h4>
            <p className="text-muted-foreground text-sm">
              An open-source collaborative knowledge graph.
            </p>
          </div>
          <Separator className="my-4" />
          <div className="text-muted-foreground flex h-5 items-center gap-4 text-sm">
            <span className="hover:text-foreground cursor-pointer transition-colors">Blog</span>
            <Separator orientation="vertical" />
            <span className="hover:text-foreground cursor-pointer transition-colors">Docs</span>
            <Separator orientation="vertical" />
            <span className="hover:text-foreground cursor-pointer transition-colors">Source</span>
          </div>
        </div>
        <ModeToggle />
      </div>
    </footer>
  );
}
