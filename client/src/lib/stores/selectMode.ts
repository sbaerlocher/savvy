import { writable } from 'svelte/store';

/**
 * True while a list screen is in batch-selection mode. Android replaces the
 * whole navigation chrome in that state (M3 contextual top app bar + batch
 * bottom bar, no nav bar and no FAB), so the layout has to know — a store
 * rather than a prop, since WalletView sits several levels below it.
 */
export const selectModeActive = writable(false);
