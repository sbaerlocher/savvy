/**
 * Creates a debounced function that delays invoking func until after delay milliseconds
 * have elapsed since the last time the debounced function was invoked.
 *
 * @param fn - The function to debounce
 * @param delay - The number of milliseconds to delay
 * @returns A debounced version of the function
 *
 * @example
 * ```typescript
 * const debouncedSearch = debounce((value: string) => {
 *   search = value;
 * }, 300);
 *
 * // In template:
 * <input on:input={(e) => debouncedSearch(e.currentTarget.value)} />
 * ```
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function debounce<T extends (...args: any[]) => void>(
	fn: T,
	delay: number
): (...args: Parameters<T>) => void {
	let timeout: ReturnType<typeof setTimeout>;

	return (...args: Parameters<T>) => {
		clearTimeout(timeout);
		timeout = setTimeout(() => fn(...args), delay);
	};
}
