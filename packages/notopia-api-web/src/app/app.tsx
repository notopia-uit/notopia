// Uncomment this line to use CSS modules
// import styles from './app.module.css';
import { ApiReferenceReact } from "@scalar/api-reference-react";
import "@scalar/api-reference-react/style.css";
import noteApiSpec from "notopia-api-note" with { type: "json" };

export function App() {
  return (
    <div>
      <ApiReferenceReact
        configuration={{
          url: "https://registry.scalar.com/@scalar/apis/galaxy?format=yaml",
          sources: [
            {
              content: noteApiSpec,
              default: true,
              title: "Note API",
            },
          ],
        }}
      />
    </div>
  );
}

export default App;
