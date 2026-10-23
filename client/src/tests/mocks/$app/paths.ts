// Mock for $app/paths
// Mirrors SvelteKit 3's `resolve` with an empty base path: route IDs start
// with `/` and get their `[param]` segments filled; pathnames have no leading
// slash and are made root-relative.
export function resolve(path: string, params?: Record<string, string>): string {
	if (!path.startsWith('/')) return `/${path}`;
	if (!params) return path;
	return path.replace(/\[(\w+)\]/g, (_, key) => params[key] ?? `[${key}]`);
}
