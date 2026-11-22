import{Ct as e,Q as t,it as n,rt as r,t as i,tt as a}from"./chunks/framework.Bol8dOGv.js";var o=JSON.parse(`{"title":"Billing Class Diagram","description":"","frontmatter":{"order":2},"headers":[],"relativePath":"services/billing/domain.md","filePath":"services/billing/domain.md"}`),s={name:`services/billing/domain.md`};function c(i,o,s,c,l,u){return e(),a(`div`,{"data-pagefind-body":!0,"data-pagefind-meta":`date:1795268755000`},[...o[0]||=[r("",4),t(`figure`,{class:`vpd-diagram vpd-diagram--plantuml`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[t(`img`,{src:`/notopia/pr/287/diagrams/plantuml-class-diagram-billing-domain-b2c6926809ff7db40a43824ed565cc14.svg`,alt:`plantuml Diagram`,class:`vpd-diagram-image`})],-1),t(`h2`,{id:`application-layer`,tabindex:`-1`},[n(`Application Layer `),t(`a`,{class:`header-anchor`,href:`#application-layer`,"aria-label":`Permalink to “Application Layer”`},`​`)],-1),t(`p`,null,[t(`code`,null,`CancelNow`),n(` only records the local cancellation and refund requests. The provider side effects are separate commands driven by billing's own outbox events (see `),t(`a`,{href:`./flows.html#cancel-now`},`Flows`),n(`).`)],-1),t(`figure`,{class:`vpd-diagram vpd-diagram--plantuml`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[t(`img`,{src:`/notopia/pr/287/diagrams/plantuml-class-diagram-billing-application-9e30067dd8d465f565fde1baf79aceef.svg`,alt:`plantuml Diagram`,class:`vpd-diagram-image`})],-1)]])}var l=i(s,[[`render`,c]]);export{o as __pageData,l as default};