'use client';

import {
  useHocuspocusAwareness,
  useHocuspocusSyncStatus,
  useHocuspocusConnectionStatus,
} from '@hocuspocus/provider-react';
import { CloudCheck, CloudUpload, RefreshCw, Wifi, WifiOff } from 'lucide-react';

import { Avatar, AvatarImage, AvatarFallback } from './shadcn/avatar';
import { Badge } from './shadcn/badge';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './shadcn/tooltip';

export function EditorStatus() {
  const connection = useHocuspocusConnectionStatus();
  const sync = useHocuspocusSyncStatus();
  const users = useHocuspocusAwareness();
  
  return (
    <TooltipProvider>
      <div className="bg-background/95 supports-backdrop-filter:bg-background/60 sticky top-0 z-50 flex items-center gap-3 border-b px-4 py-2 backdrop-blur-sm">
        <div className="flex items-center gap-2">
          {connection === 'connected' ? (
            <Badge
              variant="secondary"
              className="gap-1 border-emerald-200 bg-emerald-50 px-2 text-emerald-700"
            >
              <Wifi className="size-3.5" />
              Connected
            </Badge>
          ) : connection === 'connecting' ? (
            <Badge variant="outline" className="text-muted-foreground animate-pulse gap-1 px-2">
              <RefreshCw className="size-3.5 animate-spin" />
              Connecting
            </Badge>
          ) : (
            <Badge variant="destructive" className="gap-1 px-2">
              <WifiOff className="size-3.5" />
              Offline
            </Badge>
          )}
        </div>

        <Tooltip>
          <TooltipTrigger asChild>
            <div className="text-muted-foreground hover:text-foreground flex items-center transition-colors">
              {sync === 'syncing' ? (
                <CloudUpload className="size-5 animate-bounce" />
              ) : (
                <CloudCheck className="size-5 text-emerald-500" />
              )}
            </div>
          </TooltipTrigger>
          <TooltipContent>
            {sync === 'syncing' ? 'Saving changes...' : 'All changes saved to server'}
          </TooltipContent>
        </Tooltip>

        <div className="ml-auto flex items-center gap-2">
          <div className="mr-2 flex -space-x-2">
            {users.slice(0, 3).map((user) => (
              <Tooltip key={user.clientId}>
                <TooltipTrigger asChild>
                  <Avatar className="border-background ring-border size-8 border-2 ring-1">
                    <AvatarImage src={user.user.avatar} />
                    <AvatarFallback
                      style={{ backgroundColor: user.user.color }}
                      className="text-[10px] text-white"
                    >
                      {user.user.name.substring(0, 2).toUpperCase() ?? '??'}
                    </AvatarFallback>
                  </Avatar>
                </TooltipTrigger>
                <TooltipContent>{user.user.name}</TooltipContent>
              </Tooltip>
            ))}
          </div>

          {users.length > 3 && (
            <span className="text-muted-foreground text-xs">+{users.length - 3} more</span>
          )}
        </div>
      </div>
    </TooltipProvider>
  );
}
