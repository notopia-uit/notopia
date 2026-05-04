import { Separator } from './shadcn/separator';
import { ModeToggle } from './theme-mode-toggle';

export default function Footer() {
  return (
    <footer>
      <div className="container flex flex-col items-center justify-between space-y-4 border-t px-16 py-2 md:flex-row">
        <div>
          <div className="space-y-1">
            <h4 className="text-sm leading-none font-medium">Note Land</h4>
            <p className="text-muted-foreground text-sm">An open-source Note for user.</p>
          </div>
          <Separator className="my-4" />
          <div className="flex h-5 items-center space-x-4 text-sm">
            <div>Blog</div>
            <Separator orientation="vertical" />
            <div>Docs</div>
            <Separator orientation="vertical" />
            <div>Source</div>
          </div>
        </div>
        <div>
          <ModeToggle />
        </div>
      </div>
    </footer>
  );
}
