// Este arquivo informa ao TypeScript como lidar com importações de arquivos que não são TypeScript.
// Ao declarar módulos para esses tipos de arquivo, podemos importá-los em nossos arquivos .ts
// sem causar erros no TypeScript, enquanto deixamos nosso bundler (Parcel) processá-los.

declare module '*.css'; // Para que o TypeScript entenda importações de arquivos CSS

// Declaração para a API de HMR (Hot Module Replacement) do Parcel/Vite
interface ImportMeta {
  readonly hot?: {
    accept(callback: () => void): void;
  };
}