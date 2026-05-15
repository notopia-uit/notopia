'use client';
import { getMyWorkspacesOptions, useRenameWorkspaceMutation } from '@notopia-uit/api-gen';
import { ErrorAlert } from '@notopia-uit/ui/components/error-alert';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Label } from '@notopia-uit/ui/components/shadcn/label';
import { Separator } from '@notopia-uit/ui/components/shadcn/separator';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
import { SuccessAlert } from '@notopia-uit/ui/components/success-alert';
import { useAlert } from '@notopia-uit/ui/hooks/use-alert';
import { useQuery } from '@tanstack/react-query';
import { Trash2 } from 'lucide-react';
import Link from 'next/link';
import { useState, useEffect } from 'react';

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from './shadcn/alert-dialog';

interface GeneralSettingsProps {
  workspaceId: string;
}

//TODO: handle loading and error states
export function GeneralSettings({ workspaceId }: GeneralSettingsProps) {
  const {
    data: allWorkspaceData,
    isPending,
    isError,
    error,
  } = useQuery({
    ...getMyWorkspacesOptions({}),
  });
  const [workspaceName, setWorkspaceName] = useState('');
  useEffect(() => {
    if (allWorkspaceData) {
      const currentWorkspace = allWorkspaceData.find((ws) => ws.workspace.id === workspaceId);
      if (currentWorkspace) {
        setWorkspaceName(currentWorkspace.workspace.name);
      }
    }
  }, [allWorkspaceData, workspaceId]);

  if (isError) {
    throw error;
  }
  const { alert, showAlert } = useAlert();

  const { mutate: renameWorkspace, isPending: isRenaming } = useRenameWorkspaceMutation({
    onSuccess: (_, variables) => {
      showAlert(
        'success',
        'Workspace Renamed',
        `Workspace successfully renamed to "${variables.body.name}".`
      );
    },
    onError: (error) => {
      showAlert(
        'error',
        'Rename Failed',
        `Failed to rename workspace. ${error instanceof Error ? error.message : 'Please try again.'}`
      );
    },
  });
  return isPending ? (
    <Spinner />
  ) : (
    <div className="space-y-8">
      <div className="space-y-4">
        {alert?.type === 'success' && <SuccessAlert title={alert.title} message={alert.message} />}

        {alert?.type === 'error' && <ErrorAlert title={alert.title} message={alert.message} />}

        <div className="space-y-2">
          <Label htmlFor="workspace-name" className="text-zinc-200">
            Workspace Name
          </Label>
          <Input
            id="workspace-name"
            defaultValue={workspaceName}
            onChange={(e) => setWorkspaceName(e.target.value)}
            className="border-zinc-800 bg-zinc-900/50 text-zinc-100 focus-visible:ring-zinc-700"
          />
          <p className="text-sm text-zinc-500">
            This is the name that will be displayed on your workspace dashboard and invitations.
          </p>
        </div>
      </div>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            className="bg-zinc-100 text-zinc-900 hover:bg-zinc-200"
            disabled={isRenaming || !workspaceName.trim()}
          >
            {isRenaming ? 'Updating...' : 'Update workspace'}
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent className="border-zinc-800 bg-zinc-950 text-zinc-50">
          <AlertDialogHeader>
            <AlertDialogTitle>Rename Workspace</AlertDialogTitle>
            <AlertDialogDescription className="text-zinc-400">
              Are you sure you want to rename this workspace to "{workspaceName}"? This change will
              be visible to all members inside the workspace.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel className="border-zinc-800 bg-transparent text-zinc-100 hover:bg-zinc-800 hover:text-zinc-50">
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={() =>
                renameWorkspace({
                  path: { workspaceId: workspaceId },
                  body: { name: workspaceName },
                })
              }
              className="bg-zinc-100 text-zinc-900 hover:bg-zinc-200"
            >
              Confirm
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Separator className="bg-zinc-800" />

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
            <Trash2 className="size-4" />
            View Trash
          </Link>
        </Button>
      </div>
    </div>
  );
}
