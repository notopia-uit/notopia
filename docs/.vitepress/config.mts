import { defineConfig } from "vitepress";
import { configureDiagramsPlugin } from "vitepress-plugin-diagrams";

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
        publicPath: "/notopia/public/diagrams",
        krokiServerUrl: process.env.KROKI_SERVER_URL,
      });
    },
  },
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: "Home", link: "../" },
      { text: "Docs", link: "/docs" },
    ],

    sidebar: [
      {
        text: "Docs",
        link: "/docs",
        items: [{ text: "Class", link: "/docs/class" }],
      },
    ],

    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/notopia-uit/notopia",
      },
    ],

    search: {
      provider: "local",
    },
  },
});
