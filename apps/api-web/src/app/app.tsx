// Uncomment this line to use CSS modules
// import styles from './app.module.css';
import openapi from '@notopia-uit/api/openapi' with { type: 'json' };
import { ApiReferenceReact } from '@scalar/api-reference-react';

import '@scalar/api-reference-react/style.css';

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
          showOperationId: true,
          persistAuth: true,
          telemetry: false,
        }}
      />
    </div>
  );
}

export default App;
