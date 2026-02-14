// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {} // Not needed with ssr=false (CSR-only mode)
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}

	// Vite environment variables
	const __APP_VERSION__: string;
}

// Vite PWA Plugin virtual modules
declare module 'virtual:pwa-register' {
	export interface RegisterSWOptions {
		immediate?: boolean;
		onNeedRefresh?: () => void;
		onOfflineReady?: () => void;
		onRegistered?: (
			registration: ServiceWorkerRegistration | undefined
		) => void;
		onRegisteredSW?: (
			swScriptUrl: string,
			registration: ServiceWorkerRegistration | undefined
		) => void;
		onRegisterError?: (error: any) => void;
	}

	export function registerSW(
		options?: RegisterSWOptions
	): (reloadPage?: boolean) => Promise<void>;
}

export {};
