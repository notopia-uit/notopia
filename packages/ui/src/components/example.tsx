'use client';

import { getWorkspaceSearchTokenOptions } from '@notopia-uit/api-gen';
import { useQuery } from '@tanstack/react-query';

export default function Example() {
  const { data, error, isError, isPending } = useQuery({
    ...getWorkspaceSearchTokenOptions({
      path: {
        workspaceId: 'notopia',
      },
    }),
  });
}
