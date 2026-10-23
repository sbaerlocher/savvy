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

export {};
