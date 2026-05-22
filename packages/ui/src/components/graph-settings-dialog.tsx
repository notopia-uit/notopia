'use client';

import { D3Config } from '@notopia-uit/ui/graph-view/graph';
import { Button } from './shadcn/button';
import { Label } from './shadcn/label';
import { Input } from './shadcn/input';
import { Checkbox } from './shadcn/checkbox';
import { Separator } from './shadcn/separator';
import { ScrollArea } from './shadcn/scroll-area';
import { useState } from 'react';
import { ChevronDown, X } from 'lucide-react';

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

  const handleNumericChange = (key: keyof D3Config, value: string) => {
    const numValue = parseFloat(value) || 0;
    onSettingsChange({
      ...currentSettings,
      [key]: numValue,
    });
  };

  const handleBooleanChange = (key: keyof D3Config, value: boolean) => {
    onSettingsChange({
      ...currentSettings,
      [key]: value,
    });
  };

  if (!isOpen) return null;

  return (
    <div className="absolute top-4 right-4 z-20">
      <div className="bg-background/95 backdrop-blur-sm border rounded-lg shadow-lg w-80">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <h3 className="font-semibold text-sm">
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
                className={`size-4 transition-transform ${
                  isCollapsed ? '-rotate-90' : ''
                }`}
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

        {/* Content */}
        {!isCollapsed && (
          <ScrollArea className="h-[400px] pr-4">
            <div className="space-y-6 p-4">
              {/* Force Settings */}
              <div className="space-y-3">
                <h3 className="font-semibold text-sm">Force Configuration</h3>
                <div className="space-y-2">
                  <Label htmlFor="repelForce" className="text-xs">
                    Repel Force: {currentSettings.repelForce?.toFixed(2)}
                  </Label>
                  <Input
                    id="repelForce"
                    type="range"
                    min="0"
                    max="1"
                    step="0.01"
                    value={currentSettings.repelForce || 0}
                    onChange={(e) =>
                      handleNumericChange('repelForce', e.target.value)
                    }
                    className="h-2"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="centerForce" className="text-xs">
                    Center Force: {currentSettings.centerForce?.toFixed(2)}
                  </Label>
                  <Input
                    id="centerForce"
                    type="range"
                    min="0"
                    max="1"
                    step="0.01"
                    value={currentSettings.centerForce || 0}
                    onChange={(e) =>
                      handleNumericChange('centerForce', e.target.value)
                    }
                    className="h-2"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="linkDistance" className="text-xs">
                    Link Distance: {currentSettings.linkDistance}
                  </Label>
                  <Input
                    id="linkDistance"
                    type="range"
                    min="10"
                    max="100"
                    step="1"
                    value={currentSettings.linkDistance || 0}
                    onChange={(e) =>
                      handleNumericChange('linkDistance', e.target.value)
                    }
                    className="h-2"
                  />
                </div>
              </div>

              <Separator />

              {/* Display Settings */}
              <div className="space-y-3">
                <h3 className="font-semibold text-sm">Display Configuration</h3>

                {isLocalGraph && (
                  <div className="space-y-2">
                    <Label htmlFor="depth" className="text-xs">
                      Depth: {currentSettings.depth}
                    </Label>
                    <Input
                      id="depth"
                      type="range"
                      min="-1"
                      max="5"
                      step="1"
                      value={currentSettings.depth || 0}
                      onChange={(e) =>
                        handleNumericChange('depth', e.target.value)
                      }
                      className="h-2"
                    />
                    <p className="text-xs text-muted-foreground">
                      -1 shows all connected nodes, 0+ shows limited depth
                    </p>
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="scale" className="text-xs">
                    Scale: {currentSettings.scale?.toFixed(2)}
                  </Label>
                  <Input
                    id="scale"
                    type="range"
                    min="0.5"
                    max="2"
                    step="0.05"
                    value={currentSettings.scale || 1}
                    onChange={(e) =>
                      handleNumericChange('scale', e.target.value)
                    }
                    className="h-2"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="fontSize" className="text-xs">
                    Font Size: {currentSettings.fontSize?.toFixed(2)}
                  </Label>
                  <Input
                    id="fontSize"
                    type="range"
                    min="0.3"
                    max="1"
                    step="0.05"
                    value={currentSettings.fontSize || 0.6}
                    onChange={(e) =>
                      handleNumericChange('fontSize', e.target.value)
                    }
                    className="h-2"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="opacityScale" className="text-xs">
                    Opacity Scale: {currentSettings.opacityScale?.toFixed(2)}
                  </Label>
                  <Input
                    id="opacityScale"
                    type="range"
                    min="0"
                    max="2"
                    step="0.1"
                    value={currentSettings.opacityScale || 1}
                    onChange={(e) =>
                      handleNumericChange('opacityScale', e.target.value)
                    }
                    className="h-2"
                  />
                </div>
              </div>

              <Separator />

              {/* Interaction Settings */}
              <div className="space-y-3">
                <h3 className="font-semibold text-sm">Interaction</h3>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="drag"
                    checked={currentSettings.drag ?? true}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('drag', checked as boolean)
                    }
                  />
                  <Label
                    htmlFor="drag"
                    className="text-xs cursor-pointer"
                  >
                    Enable Drag
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="zoom"
                    checked={currentSettings.zoom ?? true}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('zoom', checked as boolean)
                    }
                  />
                  <Label
                    htmlFor="zoom"
                    className="text-xs cursor-pointer"
                  >
                    Enable Zoom
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="focusOnHover"
                    checked={currentSettings.focusOnHover ?? false}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('focusOnHover', checked as boolean)
                    }
                  />
                  <Label
                    htmlFor="focusOnHover"
                    className="text-xs cursor-pointer"
                  >
                    Focus on Hover
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="enableRadial"
                    checked={currentSettings.enableRadial ?? false}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('enableRadial', checked as boolean)
                    }
                  />
                  <Label
                    htmlFor="enableRadial"
                    className="text-xs cursor-pointer"
                  >
                    Enable Radial Layout
                  </Label>
                </div>

                <div className="flex items-center gap-2">
                  <Checkbox
                    id="showTags"
                    checked={currentSettings.showTags ?? true}
                    onCheckedChange={(checked) =>
                      handleBooleanChange('showTags', checked as boolean)
                    }
                  />
                  <Label
                    htmlFor="showTags"
                    className="text-xs cursor-pointer"
                  >
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
