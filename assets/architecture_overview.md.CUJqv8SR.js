import{Q as e,it as t,t as n,tt as r,xt as i}from"./chunks/framework.CuR1irRm.js";var a=JSON.parse(`{"title":"Overview","description":"","frontmatter":{"title":"Overview","order":1},"headers":[],"relativePath":"architecture/overview.md","filePath":"architecture/overview.md"}`),o={name:`architecture/overview.md`};function s(n,a,o,s,c,l){return i(),r(`div`,{"data-pagefind-body":!0,"data-pagefind-meta":`date:1793506373000`},[...a[0]||=[e(`h1`,{id:`system-architecture`,tabindex:`-1`},[t(`System Architecture `),e(`a`,{class:`header-anchor`,href:`#system-architecture`,"aria-label":`Permalink to “System Architecture”`},`​`)],-1),e(`p`,null,`Notopia is a set of small services behind a single gateway (Traefik). Go services (note, authorization, billing) and NestJS services (document, search worker) talk to each other over gRPC for synchronous calls and over Redpanda for events. Faded boxes are planned or unused.`,-1),e(`p`,null,[t(`See `),e(`a`,{href:`./messaging.html`},`Messaging`),t(` for the event backbone and `),e(`a`,{href:`./auth-and-gateway.html`},`Auth and Gateway`),t(` for how requests are authenticated.`)],-1),e(`figure`,{class:`vpd-diagram vpd-diagram--d2`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[e(`img`,{src:`/notopia/diagrams/d2-architecture-diagram-6bf76f01d890a8aa049eda61cc1c1a2a.svg`,alt:`d2 Diagram`,class:`vpd-diagram-image`})],-1)]])}var c=n(o,[[`render`,s]]);export{a as __pageData,c as default};