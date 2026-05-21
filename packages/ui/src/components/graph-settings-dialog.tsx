'use client';

import { useState } from 'react';
import { D3Config } from '@notopia-uit/ui/graph-view/graph';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from './shadcn/dialog';
import { Button } from './shadcn/button';
import { Label } from './shadcn/label';
import { Input } from './shadcn/input';
import { Checkbox } from './shadcn/checkbox';
import { Separator } from './shadcn/separator';
import { ScrollArea } from './shadcn/scroll-area';

interface GraphSettingsDialogProps {
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  isLocalGraph?: boolean;
  currentSettings: Partial<D3Config>;
  onSave: (settings: Partial<D3Config>) => void;
}

export function GraphSettingsDialog({
  isOpen,
  onOpenChange,
  isLocalGraph = false,
  currentSettings,
  onSave,
}: GraphSettingsDialogProps) {
  const [tempSettings, setTempSettings] = useState<Partial<D3Config>>(currentSettings);

  const handleNumericChange = (key: keyof D3Config, value: string) => {
    const numValue = parseFloat(value) || 0;
    setTempSettings((prev) => ({
      ...prev,
      [key]: numValue,
    }));
  };

  const handleBooleanChange = (key: keyof D3Config, value: boolean) => {
    setTempSettings((prev) => ({
      ...prev,
      [key]: value,
    }));
  };

  const handleSave = () => {
    onSave(tempSettings);
    onOpenChange(false);
  };

  const handleCancel = () => {
    setTempSettings(currentSettings);
    onOpenChange(false);
  };

  const handleReset = () => {
    setTempSettings(currentSettings);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>
            Graph Settings {isLocalGraph ? '(Local)' : '(Global)'}
          </DialogTitle>
        </DialogHeader>
        <ScrollArea className="h-[400px] pr-4">
          <div className="space-y-6">
            {/* Force Settings */}
            <div className="space-y-3">
              <h3 className="font-semibold text-sm">Force Configuration</h3>
              <div className="space-y-2">
                <Label htmlFor="repelForce" className="text-xs">
                  Repel Force: {tempSettings.repelForce?.toFixed(2)}
                </Label>
                <Input
                  id="repelForce"
                  type="range"
                  min="0"
                  max="1"
                  step="0.01"
                  value={tempSettings.repelForce || 0}
                  onChange={(e) => handleNumericChange('repelForce', e.target.value)}
                  className="h-2"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="centerForce" className="text-xs">
                  Center Force: {tempSettings.centerForce?.toFixed(2)}
                </Label>
                <Input
                  id="centerForce"
                  type="range"
                  min="0"
                  max="1"
                  step="0.01"
                  value={tempSettings.centerForce || 0}
                  onChange={(e) => handleNumericChange('centerForce', e.target.value)}
                  className="h-2"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="linkDistance" className="text-xs">
                  Link Distance: {tempSettings.linkDistance}
                </Label>
                <Input
                  id="linkDistance"
                  type="range"
                  min="10"
                  max="100"
                  step="1"
                  value={tempSettings.linkDistance || 0}
                  onChange={(e) => handleNumericChange('linkDistance', e.target.value)}
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
                    Depth: {tempSettings.depth}
                  </Label>
                  <Input
                    id="depth"
                    type="range"
                    min="-1"
                    max="5"
                    step="1"
                    value={tempSettings.depth || 0}
                    onChange={(e) => handleNumericChange('depth', e.target.value)}
                    className="h-2"
                  />
                  <p className="text-xs text-muted-foreground">
                    -1 shows all connected nodes, 0+ shows limited depth
                  </p>
                </div>
              )}

              <div className="space-y-2">
                <Label htmlFor="scale" className="text-xs">
                  Scale: {tempSettings.scale?.toFixed(2)}
                </Label>
                <Input
                  id="scale"
                  type="range"
                  min="0.5"
                  max="2"
                  step="0.05"
                  value={tempSettings.scale || 1}
                  onChange={(e) => handleNumericChange('scale', e.target.value)}
                  className="h-2"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="fontSize" className="text-xs">
                  Font Size: {tempSettings.fontSize?.toFixed(2)}
                </Label>
                <Input
                  id="fontSize"
                  type="range"
                  min="0.3"
                  max="1"
                  step="0.05"
                  value={tempSettings.fontSize || 0.6}
                  onChange={(e) => handleNumericChange('fontSize', e.target.value)}
                  className="h-2"
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="opacityScale" className="text-xs">
                  Opacity Scale: {tempSettings.opacityScale?.toFixed(2)}
                </Label>
                <Input
                  id="opacityScale"
                  type="range"
                  min="0"
                  max="2"
                  step="0.1"
                  value={tempSettings.opacityScale || 1}
                  onChange={(e) => handleNumericChange('opacityScale', e.target.value)}
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
                  checked={tempSettings.drag ?? true}
                  onCheckedChange={(checked) =>
                    handleBooleanChange('drag', checked as boolean)
                  }
                />
                <Label htmlFor="drag" className="text-xs cursor-pointer">
                  Enable Drag
                </Label>
              </div>

              <div className="flex items-center gap-2">
                <Checkbox
                  id="zoom"
                  checked={tempSettings.zoom ?? true}
                  onCheckedChange={(checked) =>
                    handleBooleanChange('zoom', checked as boolean)
                  }
                />
                <Label htmlFor="zoom" className="text-xs cursor-pointer">
                  Enable Zoom
                </Label>
              </div>

              <div className="flex items-center gap-2">
                <Checkbox
                  id="focusOnHover"
                  checked={tempSettings.focusOnHover ?? false}
                  onCheckedChange={(checked) =>
                    handleBooleanChange('focusOnHover', checked as boolean)
                  }
                />
                <Label htmlFor="focusOnHover" className="text-xs cursor-pointer">
                  Focus on Hover
                </Label>
              </div>

              <div className="flex items-center gap-2">
                <Checkbox
                  id="enableRadial"
                  checked={tempSettings.enableRadial ?? false}
                  onCheckedChange={(checked) =>
                    handleBooleanChange('enableRadial', checked as boolean)
                  }
                />
                <Label htmlFor="enableRadial" className="text-xs cursor-pointer">
                  Enable Radial Layout
                </Label>
              </div>

              <div className="flex items-center gap-2">
                <Checkbox
                  id="showTags"
                  checked={tempSettings.showTags ?? true}
                  onCheckedChange={(checked) =>
                    handleBooleanChange('showTags', checked as boolean)
                  }
                />
                <Label htmlFor="showTags" className="text-xs cursor-pointer">
                  Show Tags
                </Label>
              </div>
            </div>
          </div>
        </ScrollArea>

        <Separator className="mt-4" />

        <div className="flex justify-end gap-2 pt-4">
          <Button variant="outline" onClick={handleCancel}>
            Cancel
          </Button>
          <Button variant="outline" onClick={handleReset}>
            Reset
          </Button>
          <Button onClick={handleSave}>Save Settings</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
