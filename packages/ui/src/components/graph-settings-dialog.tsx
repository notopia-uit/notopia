'use client';

import { D3Config } from '@notopia-uit/ui/graph-view/graph';
import { ChevronDown, X } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

import { Button } from './shadcn/button';
import { Checkbox } from './shadcn/checkbox';
import { Label } from './shadcn/label';
import { ScrollArea } from './shadcn/scroll-area';
import { Separator } from './shadcn/separator';
import { Slider } from './shadcn/slider';

interface GraphSettingsDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  isLocalGraph?: boolean;
  currentSettings: Partial<D3Config>;
  onSettingsChange: (settings: Partial<D3Config>) => void;
}

export function GraphSettingsDialog({
  isOpen,
  onOpenChange,
  isLocalGraph = false,
  currentSettings,
  onSettingsChange,
}: GraphSettingsDialogProps) {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [localSettings, setLocalSettings] = useState<Partial<D3Config>>(currentSettings);

  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const pendingChangesRef = useRef<Partial<D3Config>>({});
  const latestParentSettingsRef = useRef<Partial<D3Config>>(currentSettings);
  const onSettingsChangeRef = useRef(onSettingsChange);

  latestParentSettingsRef.current = currentSettings;
  onSettingsChangeRef.current = onSettingsChange;

  useEffect(() => {
    setLocalSettings(currentSettings);
  }, [currentSettings]);

  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, []);

  const handleNumericChange = (key: keyof D3Config, value: number) => {
    setLocalSettings((prev) => ({ ...prev, [key]: value }));
    (pendingChangesRef.current as Record<string, unknown>)[key] = value;

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }
    debounceTimerRef.current = setTimeout(() => {
      const changes = { ...pendingChangesRef.current };
      pendingChangesRef.current = {};
      onSettingsChangeRef.current({
        ...latestParentSettingsRef.current,
        ...changes,
      });
    }, 200);
  };

  const handleBooleanChange = (key: keyof D3Config, value: boolean) => {
    onSettingsChangeRef.current({
      ...latestParentSettingsRef.current,
      [key]: value,
    });
  };

  if (!isOpen) return null;

  return (
    <div className="absolute top-4 right-4 z-20">
      <div className="bg-background/95 w-80 rounded-lg border shadow-lg backdrop-blur-sm">
        <div className="flex items-center justify-between border-b p-4">
          <h3 className="text-sm font-semibold">
            Graph Settings {isLocalGraph ? '(Local)' : '(Global)'}
          </h3>
          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsCollapsed(!isCollapsed)}
              className="size-6 p-0"
              aria-label={isCollapsed ? 'Expand' : 'Collapse'}
            >
              <ChevronDown
                className={`size-4 transition-transform ${isCollapsed ? '-rotate-90' : ''}`}
              />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onOpenChange(false)}
              className="size-6 p-0"
              aria-label="Close settings"
            >
              <X className="size-4" />
            </Button>
          </div>
        </div>

        {!isCollapsed && (
          <ScrollArea className="h-[400px] pr-4">
            <div className="space-y-6 p-4">
              <div className="space-y-3">
                <h3 className="text-sm font-semibold">Force Configuration</h3>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="repelForce" className="text-xs">
                      Repel Force
                    </Label>
                    <span className="text-muted-foreground text-xs tabular-nums">
                      {(localSettings.repelForce ?? 0).toFixed(2)}
                    </span>
                  </div>
                  <Slider
                    id="repelForce"
                    min={0}
                    max={1}
                    step={0.01}
                    value={[localSettings.repelForce ?? 0]}
                    onValueChange={([value]) => handleNumericChange('repelForce', value)}
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="centerForce" className="text-xs">
                      Center Force
                    </Label>
                    <span className="text-muted-foreground text-xs tabular-nums">
                      {(localSettings.centerForce ?? 0).toFixed(2)}
                    </span>
                  </div>
                  <Slider
                    id="centerForce"
                    min={0}
                    max={1}
                    step={0.01}
                    value={[localSettings.centerForce ?? 0]}
                    onValueChange={([value]) => handleNumericChange('centerForce', value)}
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="linkDistance" className="text-xs">
                      Link Distance
                    </Label>
                    <span className="text-muted-foreground text-xs tabular-nums">
                      {localSettings.linkDistance ?? 0}
                    </span>
                  </div>
                  <Slider
                    id="linkDistance"
                    min={10}
                    max={100}
                    step={1}
                    value={[localSettings.linkDistance ?? 0]}
                    onValueChange={([value]) => handleNumericChange('linkDistance', value)}
                  />
                </div>
              </div>

              <Separator />

              <div className="space-y-3">
                <h3 className="text-sm font-semibold">Display Configuration</h3>

                {isLocalGraph && (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label htmlFor="depth" className="text-xs">
                        Depth
                      </Label>
                      <span className="text-muted-foreground text-xs tabular-nums">
                        {localSettings.depth ?? 0}
                      </span>
                    </div>
                    <Slider
                      id="depth"
                      min={-1}
                      max={5}
                      step={1}
                      value={[localSettings.depth ?? 0]}
                      onValueChange={([value]) => handleNumericChange('depth', value)}
                    />
                    <p className="text-muted-foreground text-xs">
                      -1 shows all connected nodes, 0+ shows limited depth
                    </p>
                  </div>
                )}

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="scale" className="text-xs">
                      Scale
                    </Label>
                    <span className="text-muted-foreground text-xs tabular-nums">
                      {(localSettings.scale ?? 1).toFixed(2)}
                    </span>
                  </div>
                  <Slider
                    id="scale"
                    min={0.5}
                    max={2}
                    step={0.05}
                    value={[localSettings.scale ?? 1]}
                    onValueChange={([value]) => handleNumericChange('scale', value)}
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="fontSize" className="text-xs">
                      Font Size
                    </Label>
                    <span className="text-muted-foreground text-xs tabular-nums">
                      {(localSettings.fontSize ?? 0.6).toFixed(2)}
                    </span>
                  </div>
                  <Slider
                    id="fontSize"
                    min={0.3}
                    max={1}
                    step={0.05}
                    value={[localSettings.fontSize ?? 0.6]}
                    onValueChange={([value]) => handleNumericChange('fontSize', value)}
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="opacityScale" className="text-xs">
                      Opacity Scale
                    </Label>
                    <span className="text-muted-foreground text-xs tabular-nums">
                      {(localSettings.opacityScale ?? 1).toFixed(2)}
                    </span>
                  </div>
                  <Slider
                    id="opacityScale"
                    min={0}
                    max={2}
                    step={0.1}
                    value={[localSettings.opacityScale ?? 1]}
                    onValueChange={([value]) => handleNumericChange('opacityScale', value)}
                  />
                </div>
              </div>

              <Separator />

              <div className="space-y-3">
                <h3 className="text-sm font-semibold">Interaction</h3>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="drag"
                    checked={localSettings.drag ?? true}
                    onCheckedChange={(checked) => handleBooleanChange('drag', checked as boolean)}
                  />
                  <Label htmlFor="drag" className="cursor-pointer text-xs">
                    Enable Drag
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="zoom"
                    checked={localSettings.zoom ?? true}
                    onCheckedChange={(checked) => handleBooleanChange('zoom', checked as boolean)}
                  />
                  <Label htmlFor="zoom" className="cursor-pointer text-xs">
                    Enable Zoom
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="focusOnHover"
                    checked={localSettings.focusOnHover ?? false}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('focusOnHover', checked as boolean)
                    }
                  />
                  <Label htmlFor="focusOnHover" className="cursor-pointer text-xs">
                    Focus on Hover
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="enableRadial"
                    checked={localSettings.enableRadial ?? false}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('enableRadial', checked as boolean)
                    }
                  />
                  <Label htmlFor="enableRadial" className="cursor-pointer text-xs">
                    Enable Radial Layout
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="showTags"
                    checked={localSettings.showTags ?? true}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('showTags', checked as boolean)
                    }
                  />
                  <Label htmlFor="showTags" className="cursor-pointer text-xs">
                    Show Tags
                  </Label>
                </div>
              </div>
            </div>
          </ScrollArea>
        )}
      </div>
    </div>
  );
}
