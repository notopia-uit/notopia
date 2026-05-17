'use client';

import { Button } from '@notopia-uit/ui/components/shadcn/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@notopia-uit/ui/components/shadcn/card';
import { FileText, Settings, Users, Zap } from 'lucide-react';
import Link from 'next/link';

interface WorkspaceWelcomeProps {
  workspaceId: string;
  workspaceName: string;
  workspaceSlug: string;
}

const WorkspaceWelcome = ({ workspaceId, workspaceName, workspaceSlug }: WorkspaceWelcomeProps) => {
  const quickActionItems = [
    {
      icon: FileText,
      title: 'Create a Note',
      description: 'Start documenting your ideas',
      href: `/workspace/${workspaceSlug}/note/new`,
    },
    {
      icon: Users,
      title: 'Invite Team Members',
      description: 'Collaborate with your team',
      href: `/workspace/${workspaceSlug}/settings`,
    },
    {
      icon: Zap,
      title: 'Explore Features',
      description: 'Learn what you can do',
      href: `/workspace/${workspaceSlug}/note`,
    },
    {
      icon: Settings,
      title: 'Workspace Settings',
      description: 'Customize your workspace',
      href: `/workspace/${workspaceSlug}/settings`,
    },
  ];

  return (
    <div className="from-background via-background to-muted/20 min-h-screen bg-linear-to-br p-4 md:p-8">
      <div className="mx-auto max-w-4xl">
        <div className="mb-12 space-y-4 text-center">
          <h1 className="text-4xl font-bold tracking-tight md:text-5xl">
            Welcome to <span className="text-primary">{workspaceName}</span>
          </h1>
          <p className="text-muted-foreground text-lg">
            You're all set! Let's make the most of your workspace.
          </p>
        </div>

        <div className="mb-12 grid gap-4 sm:grid-cols-2">
          {quickActionItems.map((item) => {
            const Icon = item.icon;
            return (
              <Card
                key={item.title}
                className="hover:border-primary transition-all hover:shadow-lg"
              >
                <CardHeader className="pb-3">
                  <div className="flex items-start gap-3">
                    <div className="bg-primary/10 rounded-lg p-2">
                      <Icon className="text-primary size-5" />
                    </div>
                    <div className="flex-1">
                      <CardTitle className="text-lg">{item.title}</CardTitle>
                      <CardDescription className="mt-1">{item.description}</CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent>
                  <Link href={item.href}>
                    <Button variant="outline" size="sm" className="w-full">
                      Get Started
                    </Button>
                  </Link>
                </CardContent>
              </Card>
            );
          })}
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-xl">Workspace Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-4 sm:grid-cols-3">
                <div className="space-y-1">
                  <p className="text-muted-foreground text-sm font-medium">Workspace Name</p>
                  <p className="text-lg font-semibold">{workspaceName}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-muted-foreground text-sm font-medium">Workspace Slug</p>
                  <p className="text-lg font-semibold">/{workspaceSlug}</p>
                </div>
                <div className="space-y-1">
                  <p className="text-muted-foreground text-sm font-medium">Workspace ID</p>
                  <p className="text-muted-foreground font-mono text-sm">{workspaceId}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-xl">Getting Started Tips</CardTitle>
            </CardHeader>
            <CardContent>
              <ul className="space-y-3">
                <li className="flex gap-3">
                  <span className="text-primary mt-1 font-bold">•</span>
                  <span>Create your first note to start organizing your thoughts</span>
                </li>
                <li className="flex gap-3">
                  <span className="text-primary mt-1 font-bold">•</span>
                  <span>Invite team members to collaborate on shared documents</span>
                </li>
                <li className="flex gap-3">
                  <span className="text-primary mt-1 font-bold">•</span>
                  <span>Use the sidebar to navigate through your workspace</span>
                </li>
                <li className="flex gap-3">
                  <span className="text-primary mt-1 font-bold">•</span>
                  <span>Customize workspace settings to match your needs</span>
                </li>
              </ul>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};

export { WorkspaceWelcome };
