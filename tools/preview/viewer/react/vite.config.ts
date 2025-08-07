import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
// import prismjsPlugin from "vite-plugin-prismjs";

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    // prismjsPlugin({
    //   languages: [
    //     'bash',
    //     'batch',
    //     'css',
    //     'git',
    //     'go',
    //     'java',
    //     'javascript',
    //     'json',
    //     'json5',
    //     'makefile',
    //     'markup',
    //     'mermaid',
    //     'promql',
    //     'protobuf',
    //     'python',
    //     'reason',
    //     'sass',
    //     'scss',
    //     'tsx',
    //     'typescript',
    //     'xml-doc',
    //     'yaml'
    //   ],
    //   plugins: [
    //     'copy-to-clipboard',
    //     'show-language',
    //     /**
    //      * must be applied to an HTML tag
    //      * (ideally to <body> because line-numbers is inherited,
    //      * i.e., it's applied to all its children)
    //      */
    //     'line-numbers',
    //     /**
    //      * must be applied to an HTML tag
    //      * (ideally to <body> because match-braces is inherited,
    //      * i.e., it's applied to all its children)
    //     */
    //     'match-braces'
    //   ],
    //   /**
    //    * to apply a theme, referencing it below is enough because Vite bundles it at build time
    //    * there is no need to import the CSS file in the code directly
    //    * */
    //   theme: 'okaidia',
    //   css: true,
    // }),
  ],
  // optimizeDeps: {
  //   include: ['prismjs'],
  // },
})
