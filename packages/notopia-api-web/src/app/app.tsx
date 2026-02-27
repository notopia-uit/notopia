// Uncomment this line to use CSS modules
// import styles from './app.module.css';
import '@scalar/api-reference-react/style.css';

import { ApiReferenceReact } from '@scalar/api-reference-react';
import openapi from 'api/openapi' with { type: 'json' };

export function App() {
  return (
    <div>
      <ApiReferenceReact
        configuration={{
          sources: [
            {
              content: openapi,
              default: true,
              title: 'Notopia API',
            },
          ],
          persistAuth: true,
          telemetry: false,
        }}
      />
    </div>
  );
}

export default App;
