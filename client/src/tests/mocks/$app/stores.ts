// Mock for $app/stores
import { writable } from 'svelte/store';

export const page = writable({});
export const navigating = writable(null);
export const updated = writable(false);
