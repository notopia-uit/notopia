import{Tt as e,et as t,ot as n,rt as r,t as i}from"./chunks/framework.CZq4qJDD.js";var a=JSON.parse(`{"title":"Document Database Diagram","description":"","frontmatter":{"order":3},"headers":[],"relativePath":"services/document/data.md","filePath":"services/document/data.md"}`),o={name:`services/document/data.md`};function s(i,a,o,s,c,l){return e(),r(`div`,{"data-pagefind-body":!0,"data-pagefind-meta":`date:1795287477000`},[...a[0]||=[t(`h1`,{id:`document-database-diagram`,tabindex:`-1`},[n(`Document Database Diagram `),t(`a`,{class:`header-anchor`,href:`#document-database-diagram`,"aria-label":`Permalink to “Document Database Diagram”`},`​`)],-1),t(`div`,{class:`info custom-block`},[t(`p`,{class:`custom-block-title custom-block-title-default`},`INFO`),t(`ul`,null,[t(`li`,null,[n(`Only column marked with `),t(`code`,null,`N`),n(` mean nullable`)])])],-1),t(`figure`,{class:`vpd-diagram vpd-diagram--d2`,onclick:`
        const figure = this;
        const isFullscreen = figure.classList.contains('vpd-diagram--fullscreen');

        document.querySelectorAll('.vpd-diagram').forEach(diagram => {
          diagram.classList.remove('vpd-diagram--fullscreen');
        });

        if (!isFullscreen) {
          figure.classList.add('vpd-diagram--fullscreen');
        }
      `},[t(`img`,{src:`/notopia/diagrams/d2-database-diagram-document-6010ef3bbb33f5f2e68e32e88a1429e8.svg`,alt:`d2 Diagram`,class:`vpd-diagram-image`})],-1),t(`p`,null,[n(`Revision content is serialized JSON stored in a text column. Revisions are hard-deleted, and deleting a document cascades to its revisions. See `),t(`a`,{href:`./flows.html`},`flows`),n(` for commit and deletion behavior.`)],-1)]])}var c=i(o,[[`render`,s]]);export{a as __pageData,c as default};