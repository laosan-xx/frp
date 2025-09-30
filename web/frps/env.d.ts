/// <reference types="vite/client" />

// Minimal JSX fallback to avoid "JSX.IntrinsicElements" errors when not using TSX
declare namespace JSX {
  interface IntrinsicElements {
    [elemName: string]: any
  }
}
