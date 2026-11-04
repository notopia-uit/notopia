import{Q as e,it as t,rt as n,t as r,tt as i,xt as a}from"./chunks/framework.CRPOboOd.js";var o=JSON.parse(`{"title":"Billing Class Diagram","description":"","frontmatter":{"order":3},"headers":[],"relativePath":"services/billing/domain.md","filePath":"services/billing/domain.md"}`),s={name:`services/billing/domain.md`};function c(r,o,s,c,l,u){return a(),i(`div`,{"data-pagefind-body":!0,"data-pagefind-meta":`date:1793531709000`},[...o[0]||=[n("",5),e(`figure`,{class:`vpd-diagram vpd-diagram--plantuml`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[e(`img`,{src:`/notopia/diagrams/plantuml-class-diagram-billing-domain-8bd100af0690f4eb8f5ca0d7baa3dd74.svg`,alt:`plantuml Diagram`,class:`vpd-diagram-image`})],-1),e(`h2`,{id:`application-layer`,tabindex:`-1`},[t(`Application Layer `),e(`a`,{class:`header-anchor`,href:`#application-layer`,"aria-label":`Permalink to “Application Layer”`},`​`)],-1),e(`p`,null,[e(`code`,null,`CancelNow`),t(` only records the local cancellation and refund requests. The provider side effects are separate commands driven by billing's own outbox events (see `),e(`a`,{href:`./flows.html#cancel-now`},`Flows`),t(`).`)],-1),e(`figure`,{class:`vpd-diagram vpd-diagram--plantuml`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[e(`img`,{src:`/notopia/diagrams/plantuml-class-diagram-billing-application-13a09c99a3478faf991dad7bf3a3d2da.svg`,alt:`plantuml Diagram`,class:`vpd-diagram-image`})],-1)]])}var l=r(s,[[`render`,c]]);export{o as __pageData,l as default};