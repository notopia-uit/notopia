import{Q as e,rt as t,t as n,tt as r,xt as i}from"./chunks/framework.CuR1irRm.js";var a=JSON.parse(`{"title":"Note","description":"","frontmatter":{"title":"Note","order":1},"headers":[],"relativePath":"services/note/index.md","filePath":"services/note/index.md"}`),o={name:`services/note/index.md`};function s(n,a,o,s,c,l){return i(),r(`div`,{"data-pagefind-body":!0,"data-pagefind-meta":`date:1793506373000`},[...a[0]||=[t(`<h1 id="note-service" tabindex="-1">Note Service <a class="header-anchor" href="#note-service" aria-label="Permalink to “Note Service”">​</a></h1><p>Go service owning workspaces, folders, notes, links, tags, trash and publishing. DDD + CQRS, with a transactional outbox for integration events. Note only stores note metadata; the editable binary content lives in the <a href="./../document/">Document service</a>.</p><ul><li><a href="./domain.html">Domain</a>: class diagrams</li><li><a href="./data.html">Data</a>: database schema</li><li><a href="./flows.html">Flows</a>: sequence diagrams</li><li><a href="./events.html">Events</a>: topics published and consumed</li></ul><h2 id="architecture" tabindex="-1">Architecture <a class="header-anchor" href="#architecture" aria-label="Permalink to “Architecture”">​</a></h2>`,4),e(`figure`,{class:`vpd-diagram vpd-diagram--d2`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[e(`img`,{src:`/notopia/diagrams/d2-architecture-note-diagram-3b4c62eafe7cabdf266e2022524a11e1.svg`,alt:`d2 Diagram`,class:`vpd-diagram-image`})],-1)]])}var c=n(o,[[`render`,s]]);export{a as __pageData,c as default};