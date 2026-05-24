import { NoteWorkspaceRole } from '@notopia-uit/api-gen';
import { SelectItem } from '@notopia-uit/ui/components/shadcn/select';

type UserRole = typeof NoteWorkspaceRole[keyof typeof NoteWorkspaceRole];

const ROLE_DISPLAY_NAMES: Record<UserRole, string> = {
  owner: 'Owner',
  editor: 'Editor',
  viewer: 'Viewer',
};

interface RoleSelectItemsProps {
  excludeRoles?: UserRole[];
}

export function RoleSelectItems({ excludeRoles = [] }: RoleSelectItemsProps) {
  const roles = Object.values(NoteWorkspaceRole) as UserRole[];

  return (
    <>
      {roles
        .filter((role) => !excludeRoles.includes(role))
        .map((role) => (
          <SelectItem key={role} value={role}>
            {ROLE_DISPLAY_NAMES[role]}
          </SelectItem>
        ))}
    </>
  );
}
