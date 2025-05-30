declare global {
  interface Window {
    env: {
      production?: boolean;
      apiUrl?: string;
      debug?: boolean;
    };
  }
}

export const environment = {
  production: window.env?.production ?? true,
  apiUrl: window.env?.apiUrl ?? "default",
  debug: window.env?.debug ?? false,
};