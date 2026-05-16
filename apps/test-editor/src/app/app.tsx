import '@blocknote/core/fonts/inter.css';
import { useCreateBlockNote } from '@blocknote/react';
import { BlockNoteView } from '@blocknote/shadcn';
import { HocuspocusProviderWebsocketComponent, HocuspocusRoom } from '@hocuspocus/provider-react';

import '@blocknote/shadcn/style.css';
import { useHocuspocusProvider, useHocuspocusConnectionStatus } from '@hocuspocus/provider-react';
import { createBlockNoteSchema } from '@notopia-uit/ui';
import type { User } from 'oidc-client-ts';
import { useMemo } from 'react';
import { useAuth } from 'react-oidc-context';

function getDeterministicColor(id: string): string {
  let hash = 0;
  for (let i = 0; i < id.length; i++) {
    hash = id.charCodeAt(i) + ((hash << 5) - hash);
  }
  const hue = Math.abs(hash) % 360;
  return `hsl(${hue}, 65%, 55%)`;
}

function AuthenticatedEditor({ user }: { user: User }) {
  const schema = useMemo(() => createBlockNoteSchema(), []);
  const provider = useHocuspocusProvider();
  const connectionStatus = useHocuspocusConnectionStatus();
  const editor = useCreateBlockNote({
    schema,
    collaboration: {
      provider: {
        awareness: provider.awareness ?? undefined,
      },
      fragment: provider.document.getXmlFragment('document-store'),
      user: {
        name: user.profile?.name ?? user.profile?.preferred_username ?? 'Anonymous',
        color: getDeterministicColor(user.profile?.sub ?? 'anonymous'),
        avatar: user.profile?.picture ?? 'https://placehold.net/default.svg',
      },
    },
  });

  console.log('Connection status:', connectionStatus);
  if (connectionStatus !== 'connected') {
    return <LoadingSpinner />;
  }

  return (
    <BlockNoteView
      editor={editor}
      onInvalid={(e) => {
        console.error('BlockNoteView error:', e);
      }}
    />
  );
}

function SignInButton() {
  const auth = useAuth();
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
        flexDirection: 'column',
        gap: '1rem',
      }}
    >
      <p>You need to sign in to use the editor.</p>
      <button onClick={() => auth.signinRedirect()}>Sign In</button>
    </div>
  );
}

function LoadingSpinner() {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
      }}
    >
      <p>Loading...</p>
    </div>
  );
}

function ErrorMessage({ message }: { message: string }) {
  return (
    <div
      style={{
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        minHeight: '100vh',
      }}
    >
      <p style={{ color: 'red' }}>{message}</p>
    </div>
  );
}

export function App() {
  const auth = useAuth();
  const hocuspocusUrl = import.meta.env.VITE_HOCUSPOCUS_URL ?? 'ws://127.0.0.1:8082';
  const noteId = import.meta.env.VITE_NOTE_ID ?? '000c2575-1847-5293-8117-2415a2328ef8';

  if (auth.isLoading) {
    console.log('Auth loading...');
    return <LoadingSpinner />;
  }

  if (auth.error) {
    return <ErrorMessage message={`Auth error: ${auth.error.message}`} />;
  }

  if (!auth.isAuthenticated || !auth.user) {
    return <SignInButton />;
  }

  return (
    <div>
      <HocuspocusProviderWebsocketComponent url={hocuspocusUrl}>
        <HocuspocusRoom
          name={noteId}
          token={auth.user.access_token}
          onOpen={(data) => {
            console.log('Room opened', data);
          }}
          onConnect={() => {
            console.log('Connected to room');
          }}
          onDisconnect={(data) => {
            console.log('Disconnected from room', data);
          }}
          onClose={(data) => {
            console.log('Room closed', data);
          }}
          onAuthenticated={() => {
            console.log('Authenticated');
          }}
          onSynced={(data) => {
            console.log('Synced', data);
          }}
          onAuthenticationFailed={(data) => {
            console.log('Authentication failed', data);
          }}
          onDestroy={() => {
            console.log('Room destroyed');
          }}
        >
          <AuthenticatedEditor user={auth.user} />
        </HocuspocusRoom>
      </HocuspocusProviderWebsocketComponent>
    </div>
  );
}

export default App;
