import { defineConfig } from "vitepress";
import { configureDiagramsPlugin } from "vitepress-plugin-diagrams";
import { pagefindPlugin } from "vitepress-plugin-pagefind";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Notopia",
  description: "Utopia of Notes",
  lang: "en-GB",
  base: "/notopia/",
  markdown: {
    theme: {
      light: "catppuccin-latte",
      dark: "catppuccin-mocha",
    },
    config: (md) => {
      configureDiagramsPlugin(md, {
        diagramsDir: "public/diagrams",
        publicPath: "/notopia/diagrams",
        krokiServerUrl: process.env.CI
          ? undefined
          : process.env.KROKI_SERVER_URL,
      });
    },
  },
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: "Home", link: "../" },
      { text: "Docs", link: "/docs" },
      {
        text: "Scalar API",
        link: "/api/index.html",
        target: "_blank",
        rel: "noopener",
      },
    ],

    sidebar: [
      {
        text: "Docs",
        link: "/docs",
        items: [
          { text: "Class", link: "/docs/class" },
          { text: "API", link: "/docs/api" },
        ],
      },
    ],

    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/notopia-uit/notopia",
      },
    ],
  },
  vite: {
    plugins: [pagefindPlugin()],
  },
  ignoreDeadLinks: ["/api/index"],
});
