import{Tt as e,at as t,et as n,ot as r,rt as i,t as a}from"./chunks/framework.CZq4qJDD.js";var o=JSON.parse(`{"title":"Billing Class Diagram","description":"","frontmatter":{"order":2},"headers":[],"relativePath":"services/billing/domain.md","filePath":"services/billing/domain.md"}`),s={name:`services/billing/domain.md`};function c(a,o,s,c,l,u){return e(),i(`div`,{"data-pagefind-body":!0,"data-pagefind-meta":`date:1795268755000`},[...o[0]||=[t("",4),n(`figure`,{class:`vpd-diagram vpd-diagram--plantuml`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[n(`img`,{src:`/notopia/diagrams/plantuml-class-diagram-billing-domain-b2c6926809ff7db40a43824ed565cc14.svg`,alt:`plantuml Diagram`,class:`vpd-diagram-image`})],-1),n(`h2`,{id:`application-layer`,tabindex:`-1`},[r(`Application Layer `),n(`a`,{class:`header-anchor`,href:`#application-layer`,"aria-label":`Permalink to “Application Layer”`},`​`)],-1),n(`p`,null,[n(`code`,null,`CancelNow`),r(` only records the local cancellation and refund requests. The provider side effects are separate commands driven by billing's own outbox events (see `),n(`a`,{href:`./flows.html#cancel-now`},`Flows`),r(`).`)],-1),n(`figure`,{class:`vpd-diagram vpd-diagram--plantuml`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[n(`img`,{src:`/notopia/diagrams/plantuml-class-diagram-billing-application-9e30067dd8d465f565fde1baf79aceef.svg`,alt:`plantuml Diagram`,class:`vpd-diagram-image`})],-1)]])}var l=a(s,[[`render`,c]]);export{o as __pageData,l as default};