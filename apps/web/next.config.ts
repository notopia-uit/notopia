//@ts-check
import { composePlugins, withNx } from '@nx/next';
import { WithNxOptions } from '@nx/next/plugins/with-nx';

const nextConfig: WithNxOptions = {
  // Use this to set Nx-specific options
  // See: https://nx.dev/recipes/next/next-config-setup
  nx: {},
  output: 'standalone',
  allowedDevOrigins: ['web.notopia.localhost'],
  productionBrowserSourceMaps: true,
};

const plugins = [withNx];

export default composePlugins(...plugins)(nextConfig);
