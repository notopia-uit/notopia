'use client';
import { getMyWorkspacesOptions, useRenameWorkspaceMutation } from '@notopia-uit/api-gen';
import { QueryErrorFallback } from '@notopia-uit/ui/hooks/query-error-fallback';
import { useQueryErrorHandler } from '@notopia-uit/ui/hooks/use-query-error-handler';
import { Button } from '@notopia-uit/ui/components/shadcn/button';
import { Input } from '@notopia-uit/ui/components/shadcn/input';
import { Label } from '@notopia-uit/ui/components/shadcn/label';
import { Separator } from '@notopia-uit/ui/components/shadcn/separator';
import { Spinner } from '@notopia-uit/ui/components/shadcn/spinner';
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

export function GeneralSettings({ workspaceId }: GeneralSettingsProps) {
  const {
    data: allWorkspaceData,
    isPending,
    isError,
    error,
  } = useQuery({
    ...getMyWorkspacesOptions({}),
  });
  const { retry } = useQueryErrorHandler();
  const [workspaceName, setWorkspaceName] = useState('');
  useEffect(() => {
    if (allWorkspaceData) {
      const currentWorkspace = allWorkspaceData.find((ws) => ws.workspace.id === workspaceId);
      if (currentWorkspace) {
        setWorkspaceName(currentWorkspace.workspace.name);
      }
    }
  }, [allWorkspaceData, workspaceId]);

  const { showAlert } = useAlert();

  const { mutate: renameWorkspace, isPending: isRenaming } = useRenameWorkspaceMutation({
    onSuccess: (_, variables) => {
      showAlert({
        type: 'success',
        title: 'Workspace Renamed',
        message: `Workspace successfully renamed to "${variables.body.name}".`,
      });
    },
    onError: (error) => {
      showAlert({
        type: 'error',
        title: 'Rename Failed',
        message: `Failed to rename workspace. ${error instanceof Error ? error.message : 'Please try again.'}`,
      });
    },
  });
  if (isPending) {
    return <Spinner />;
  }

  if (isError) {
    return (
      <div className="space-y-4">
        <QueryErrorFallback
          error={error}
          onRetry={retry}
          title="Failed to Load Workspace Settings"
          description="Unable to fetch workspace information. Please try again."
        />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="workspace-name">
            Workspace Name
          </Label>
          <Input
            id="workspace-name"
            defaultValue={workspaceName}
            onChange={(e) => setWorkspaceName(e.target.value)}
          />
          <p className="text-sm text-muted-foreground">
            This is the name that will be displayed on your workspace dashboard and invitations.
          </p>
        </div>
      </div>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <Button
            disabled={isRenaming || !workspaceName.trim()}
          >
            {isRenaming ? 'Updating...' : 'Update workspace'}
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Rename Workspace</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to rename this workspace to "{workspaceName}"? This change will
              be visible to all members inside the workspace.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={() =>
                renameWorkspace({
                  path: { workspaceId: workspaceId },
                  body: { name: workspaceName },
                })
              }
            >
              Confirm
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      <Separator />

      <div className="space-y-4">
        <div className="space-y-2">
          <h3 className="text-lg font-medium">Recycle Bin</h3>
          <p className="text-sm text-muted-foreground">
            View and restore recently deleted notes, diagrams, and files from this workspace. Items
            remain in the trash for 30 days before permanent deletion.
          </p>
        </div>

        <Button
          variant="outline"
          className="flex items-center gap-2"
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
