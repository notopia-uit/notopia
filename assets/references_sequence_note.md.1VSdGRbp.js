import{ah as a,b as n,e as i,p as e}from"./chunks/framework.CFlXhPow.js";const D=JSON.parse('{"title":"Note","description":"","frontmatter":{},"headers":[],"relativePath":"references/sequence/note.md","filePath":"references/sequence/note.md"}'),t={name:"references/sequence/note.md"};function l(p,s,h,k,r,c){return n(),i("div",{"data-pagefind-body":!0,"data-pagefind-meta":"date:1777546875000"},[...s[0]||(s[0]=[e(`<h1 id="note" tabindex="-1">Note <a class="header-anchor" href="#note" aria-label="Permalink to “Note”">​</a></h1><div class="info custom-block"><p class="custom-block-title custom-block-title-default">INFO</p><ul><li><code>Note</code> only store the info</li><li><code>Document</code> is the <code>Note</code> content in binary format for editing and collaboration</li></ul></div><h2 id="create-note" tabindex="-1">Create Note <a class="header-anchor" href="#create-note" aria-label="Permalink to “Create Note”">​</a></h2><div class="language-mermaid"><button title="Copy Code" class="copy"></button><span class="lang">mermaid</span><pre class="shiki shiki-themes catppuccin-latte catppuccin-mocha" style="--shiki-light:#4c4f69;--shiki-dark:#cdd6f4;--shiki-light-bg:#eff1f5;--shiki-dark-bg:#1e1e2e;" tabindex="0" dir="ltr"><code><span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">sequenceDiagram</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    autonumber</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    actor U as User</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant NS as Note Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant AS as Authorization Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant MB as Message Broker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SW as Search Worker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SS as Search Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    actor SU as Subcribed Users</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    U-&gt;&gt;+NS: CreateNote</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;+AS: HasWorkspacePermission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    break Authorization Failed</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        AS--&gt;&gt;NS: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS--&gt;&gt;U: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    AS--&gt;&gt;-NS: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;NS: Create Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    par Response</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS--&gt;&gt;-U: Note ID</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    and NoteCreated domain event</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        loop Domain NoteCreatedEvent not published</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            NS-)MB: Publish domain NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        MB-)+NS: Domain NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        par Publish workspace event</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            loop Not processed domain NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">                NS-&gt;&gt;NS: Convert to workspace NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            NS-)SU: Publish workspace NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        and Publish NoteCreatedEvent integration event</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            loop Not processed domain NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">                NS-&gt;&gt;NS: Convert to NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            NS-)-MB: Publish integration NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            MB-)+SW: Retrieve integration NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            SW-&gt;&gt;SW: Process NoteCreatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            SW-&gt;&gt;+SS: Index Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            par Ack</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">                SS--&gt;&gt;-SW: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            and Indexing Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">                SS-&gt;&gt;+SS: Index Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span></code></pre></div><h2 id="get-note" tabindex="-1">Get Note <a class="header-anchor" href="#get-note" aria-label="Permalink to “Get Note”">​</a></h2><div class="language-mermaid"><button title="Copy Code" class="copy"></button><span class="lang">mermaid</span><pre class="shiki shiki-themes catppuccin-latte catppuccin-mocha" style="--shiki-light:#4c4f69;--shiki-dark:#cdd6f4;--shiki-light-bg:#eff1f5;--shiki-dark-bg:#1e1e2e;" tabindex="0" dir="ltr"><code><span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">sequenceDiagram</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    autonumber</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    actor U as User</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant NS as Note Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant DS as Document Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant AS as Authorization Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    actor SU as Subcribed Users</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    U-&gt;&gt;+NS: GetNote</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;NS: Get Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;+AS: HasNotePermission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    break Authorization Failed</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        AS--&gt;&gt;NS: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS--&gt;&gt;U: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    AS--&gt;&gt;-NS: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS--&gt;&gt;-U: Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    U-&gt;&gt;+DS: Hocuspocus Connection</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    DS-&gt;&gt;DS: Get Document</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    alt Document does not exists</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        DS-&gt;&gt;+NS: CheckNoteExistence</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS-&gt;&gt;NS: Check Note existence</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        alt Note does not exists</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            NS--&gt;&gt;-DS: Not found</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            DS--&gt;&gt;U: Not found</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        else Note exists</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            DS-&gt;&gt;DS: Init document</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    DS--&gt;&gt;U: Hocuspocus Connection</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    loop Edit debounce by handled by Hocuspocus</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        U-&gt;&gt;DS: Edit Document</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        DS-&gt;&gt;DS: Save Document changes</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        DS--&gt;&gt;SU: Broadcast Document changes</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span></code></pre></div><h2 id="commit-document" tabindex="-1">Commit Document <a class="header-anchor" href="#commit-document" aria-label="Permalink to “Commit Document”">​</a></h2><div class="language-mermaid"><button title="Copy Code" class="copy"></button><span class="lang">mermaid</span><pre class="shiki shiki-themes catppuccin-latte catppuccin-mocha" style="--shiki-light:#4c4f69;--shiki-dark:#cdd6f4;--shiki-light-bg:#eff1f5;--shiki-dark-bg:#1e1e2e;" tabindex="0" dir="ltr"><code><span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">sequenceDiagram</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    autonumber</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    actor U as User</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant DS as Document Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant MB as Message Broker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant NS as Note Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SW as Search Worker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SS as Search Service</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    U-&gt;&gt;+DS: CommitDocument</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    DS-&gt;&gt;+DS: Create Document revision</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    par Response</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        DS--&gt;&gt;-U: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    and Publish integration DocumentCommittedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        DS-)MB: Publish integration DocumentCommittedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    par Note service update note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        loop Not processed integration DocumentCommittedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            MB-)NS: Receive DocumentCommittedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS-&gt;&gt;NS: Update note size, tags</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    and Search service update note index</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        loop Not processed integration DocumentCommittedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            MB-)SW: Receive DocumentCommittedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        SW-&gt;&gt;SW: Convert to markdownContent, tags in form of NoteSearch</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        SW-&gt;&gt;+SS: Index Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        par Ack</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            SS--&gt;&gt;-SW: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        and Indexing Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            SS-&gt;&gt;+SS: Index Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span></code></pre></div><h2 id="update-note" tabindex="-1">Update Note <a class="header-anchor" href="#update-note" aria-label="Permalink to “Update Note”">​</a></h2><div class="info custom-block"><p class="custom-block-title custom-block-title-default">INFO</p><ul><li>This also include the trash and restore note</li></ul></div><div class="language-mermaid"><button title="Copy Code" class="copy"></button><span class="lang">mermaid</span><pre class="shiki shiki-themes catppuccin-latte catppuccin-mocha" style="--shiki-light:#4c4f69;--shiki-dark:#cdd6f4;--shiki-light-bg:#eff1f5;--shiki-dark-bg:#1e1e2e;" tabindex="0" dir="ltr"><code><span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">sequenceDiagram</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    autonumber</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    actor U as User</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant NS as Note Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant AS as Authorization Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant MB as Message Broker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SW as Search Worker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SS as Search Service</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    U-&gt;&gt;+NS: UpdateNote</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;+AS: HasNotePermission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    break Authorization Failed</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        AS--&gt;&gt;NS: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS--&gt;&gt;U: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    AS--&gt;&gt;-NS: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;NS: Update Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    par Response</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS--&gt;&gt;-U: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    and Publish integration NoteUpdatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        loop Not published integration NoteUpdatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            NS-)MB: Publish NoteUpdatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        loop Not processed integration NoteUpdatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            MB-)SW: NoteUpdatedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        SW-&gt;&gt;+SS: Index Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        par Ack</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            SS--&gt;&gt;-SW: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        and Indexing Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">            SS-&gt;&gt;SS: Index Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span></code></pre></div><h2 id="permanently-delete-note" tabindex="-1">Permanently Delete Note <a class="header-anchor" href="#permanently-delete-note" aria-label="Permalink to “Permanently Delete Note”">​</a></h2><div class="language-mermaid"><button title="Copy Code" class="copy"></button><span class="lang">mermaid</span><pre class="shiki shiki-themes catppuccin-latte catppuccin-mocha" style="--shiki-light:#4c4f69;--shiki-dark:#cdd6f4;--shiki-light-bg:#eff1f5;--shiki-dark-bg:#1e1e2e;" tabindex="0" dir="ltr"><code><span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">sequenceDiagram</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    autonumber</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    actor U as User</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant NS as Note Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant AS as Authorization Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant MB as Message Broker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant DS as Document Service</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SW as Search Worker</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    participant SS as Search Service</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    U-&gt;&gt;+NS: PermanentlyDeleteWorkspaceItems (note)</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;+AS: HasWorkspacePermission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    break Authorization Failed</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        AS--&gt;&gt;NS: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS--&gt;&gt;U: No Permission</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    NS-&gt;&gt;NS: Permanently delete Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    par Response</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS--&gt;&gt;-U: Ok</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    and Publish NoteDeletedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        NS-&gt;&gt;MB: Publish NoteDeletedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    par Process NoteDeletedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        MB-&gt;&gt;DS: NoteDeletedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        DS-&gt;&gt;DS: Delete Document and Revisions</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    and Search</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        MB-&gt;&gt;SW: NoteDeletedEvent</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        SW-&gt;&gt;+SS: Delete Note from index</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">        SS-&gt;&gt;+SS: Delete Note</span></span>
<span class="line"><span style="--shiki-light:#4C4F69;--shiki-dark:#CDD6F4;">    end</span></span></code></pre></div>`,13)])])}const C=a(t,[["render",l]]);export{D as __pageData,C as default};
