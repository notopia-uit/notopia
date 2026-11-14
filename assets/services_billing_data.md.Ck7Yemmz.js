import{Q as e,rt as t,t as n,tt as r,xt as i}from"./chunks/framework.CuR1irRm.js";var a=JSON.parse(`{"title":"Billing Database Diagram","description":"","frontmatter":{"order":4},"headers":[],"relativePath":"services/billing/data.md","filePath":"services/billing/data.md"}`),o={name:`services/billing/data.md`};function s(n,a,o,s,c,l){return i(),r(`div`,{"data-pagefind-body":!0,"data-pagefind-meta":`date:1793531709000`},[...a[0]||=[t(`<h1 id="billing-database-diagram" tabindex="-1">Billing Database Diagram <a class="header-anchor" href="#billing-database-diagram" aria-label="Permalink to “Billing Database Diagram”">​</a></h1><div class="info custom-block"><p class="custom-block-title custom-block-title-default">INFO</p><ul><li>Only column marked with <code>N</code> mean nullable</li><li>The Watermill outbox table (<code>outbox_messages</code>) is omitted</li><li><code>workspace_id</code> is a logical reference to the Note service workspace (no cross-service FK)</li></ul></div>`,2),e(`figure`,{class:`vpd-diagram vpd-diagram--d2`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[e(`img`,{src:`/notopia/diagrams/d2-database-diagram-billing-bbeaaabe569aa160b0478bf9663cf3b4.svg`,alt:`d2 Diagram`,class:`vpd-diagram-image`})],-1)]])}var c=n(o,[[`render`,s]]);export{a as __pageData,c as default};