import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Label } from '@notopia-uit/ui/components/shadcn/label';
import { Separator } from '@notopia-uit/ui/components/shadcn/separator';
import { Trash2 } from 'lucide-react';
import Link from 'next/link';

interface GeneralSettingsProps {
  workspaceId: string;
}

export function GeneralSettings({ workspaceId }: GeneralSettingsProps) {
  return (
    <div className="space-y-8">
      {/* Rename Workspace Form */}
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="workspace-name" className="text-zinc-200">
            Workspace Name
          </Label>
          <Input
            id="workspace-name"
            defaultValue="Notopia"
            className="border-zinc-800 bg-zinc-900/50 text-zinc-100 focus-visible:ring-zinc-700"
          />
          <p className="text-sm text-zinc-500">
            This is the name that will be displayed on your workspace dashboard and invitations.
          </p>
        </div>

        <Button className="bg-zinc-100 text-zinc-900 hover:bg-zinc-200">Update workspace</Button>
      </div>

      <Separator className="bg-zinc-800" />

      {/* Trash / Data Management Section */}
      <div className="space-y-4">
        <div className="space-y-2">
          <h3 className="text-lg font-medium text-zinc-200">Recycle Bin</h3>
          <p className="text-sm text-zinc-500">
            View and restore recently deleted notes, diagrams, and files from this workspace. Items
            remain in the trash for 30 days before permanent deletion.
          </p>
        </div>

        <Button
          variant="outline"
          className="flex items-center gap-2 border-zinc-800 text-zinc-300 hover:bg-zinc-900 hover:text-zinc-50"
          asChild
        >
          <Link href={`/workspace/${workspaceId}/trash`}>
            <Trash2 className="h-4 w-4" />
            View Trash
          </Link>
        </Button>
      </div>
    </div>
  );
}
