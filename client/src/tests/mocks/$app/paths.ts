// Mock for $app/paths
// Mirrors SvelteKit's `resolve`: fills `[param]` segments from the params
// object (empty base path in this project), otherwise returns the path as-is.
export function resolve(path: string, params?: Record<string, string>): string {
	if (!params) return path;
	return path.replace(/\[(\w+)\]/g, (_, key) => params[key] ?? `[${key}]`);
}
