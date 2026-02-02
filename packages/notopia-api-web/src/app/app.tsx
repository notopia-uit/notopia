// Uncomment this line to use CSS modules
// import styles from './app.module.css';
import { ApiReferenceReact } from "@scalar/api-reference-react";
import "@scalar/api-reference-react/style.css";
import noteApiSpec from "@notopia/note-api" with { type: "json" };

export function App() {
  return (
    <div>
      <ApiReferenceReact
        configuration={{
          sources: [
            {
              content: noteApiSpec,
              default: true,
              title: "Note API",
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
