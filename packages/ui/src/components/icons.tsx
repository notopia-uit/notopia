import { SiGithub, SiFacebook } from '@icons-pack/react-simple-icons';
import {
  LucideIcon,
  LucideProps,
  NotepadTextDashedIcon,
  SaveIcon,
  SettingsIcon,
} from 'lucide-react';

export type Icon = LucideIcon;

export const Icons = {
  Logo: NotepadTextDashedIcon,
  Github: ({ ...props }: LucideProps) => <SiGithub {...props} />,
  Facebook: ({ ...props }: LucideProps) => <SiFacebook {...props} />,
  Save: SaveIcon,
  Settings: SettingsIcon,
};
