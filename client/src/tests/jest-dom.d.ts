// Vitest 5 no longer reads matcher types from the global `jest.Matchers`
// namespace, and `@testing-library/jest-dom/vitest` still augments the
// single-parameter `Assertion<T>` that Vitest 5 replaced with
// `Assertion<R, T>`. Augment `vitest.Matchers` directly until jest-dom ships
// Vitest 5 types.
import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers';

declare module 'vitest' {
	// eslint-disable-next-line @typescript-eslint/no-empty-object-type, @typescript-eslint/no-unused-vars
	interface Matchers<R, T> extends TestingLibraryMatchers<unknown, R> {}
}
