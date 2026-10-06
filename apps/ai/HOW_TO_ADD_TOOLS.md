# How to add custom tools

Custom tools live in `src/ai/tools/` and are merged with BlockNote's own
document-editing tools on every chat request (`AiService.streamBlockNoteChat`).

## Add a tool in 3 steps

1. **Provide the data source** as a flat top-level NestJS module (`src/note/`,
   `src/authorization/`, `src/search/` — never nested inside `src/ai/`).
   Existing examples:
   - `src/note/` — gRPC client for the note service, same pattern as
     `apps/document/src/note/` (`ClientsModule.registerAsync` +
     `ClientGrpc.getService(...)` + `firstValueFrom(...)`).
   - `src/authorization/` — gRPC client for the authorization service.
   - `src/search/` — plain `Meilisearch` HTTP client.
2. **Declare the tool** in `src/ai/tools/custom-tools.ts` with `tool()` from
   the `ai` package:

   ```ts
   myTool: tool({
     description: 'What it does, when the model should call it.',
     inputSchema: z.object({
       someId: z.string().describe('What this field is.'),
     }),
     execute: async ({ someId }) => deps.myClient.fetch({ userId: deps.userId, someId }),
   }),
   ```

   Anything JSON-serializable returned from `execute` is sent back to the
   model. Return `null` (not throw) for "not found" — throwing reads as a
   system error to the model.

3. **Thread dependencies through `CustomToolDeps`** (`userId` is already
   there — the controller resolves it from `@ReqUser()` so tools always act
   as the calling user).

No registration step: `createCustomTools()` output is spread into `streamText`
next to `toolDefinitionsToToolSet(...)`.

## Rules worth knowing

- `toolChoice` is `'auto'` (not `'required'`) so the model can pick BlockNote
  tools, custom tools, or plain text. `stopWhen: stepCountIs(5)` lets it chain
  calls (e.g. `searchNotes` → `getNote`) up to 5 steps.
- Keep `execute` fast and scoped to the user: always pass `userId` into
  downstream calls (gRPC `GetNote`, Meilisearch filters).
- New external URL/key? Add it to `config.ts` (`servicesConfigSchema`) +
  `env.validation.ts` + `.env`, following the existing `NOTOPIA_AI_*` names.
